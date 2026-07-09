package logic

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	milvusDAO "techmind/internal/dao/milvus"
	redisDAO "techmind/internal/dao/redis"
	aiEmbed "techmind/internal/ai/embedding"
	aiSkill "techmind/internal/ai/prompt"
	"techmind/internal/model"
	"techmind/internal/monitor"
	"techmind/internal/pkg/settings"
	"techmind/internal/pkg/snowflake"
	"techmind/internal/worker"
)

var ErrArticleNotExist = errors.New("article not found")
var ErrForbidden = errors.New("forbidden")

// CreateArticleInput 发布文章参数
type CreateArticleInput struct {
	Title   string
	Content string
	Cover   string
	Tags    []string // 手动标签名列表
}

// UpdateArticleInput 编辑文章参数
type UpdateArticleInput struct {
	Title   string
	Content string
	Cover   string
}

// CreateArticle 发布文章
func CreateArticle(ctx context.Context, authorID int64, in *CreateArticleInput) (int64, error) {
	a := &model.Article{
		ID:       snowflake.GenID(),
		AuthorID: authorID,
		Title:    in.Title,
		Content:  in.Content,
		Cover:    in.Cover,
		Status:   1,
	}
	if err := mysqlDAO.CreateArticle(a); err != nil {
		return 0, fmt.Errorf("create article: %w", err)
	}

	// 处理手动标签
	if len(in.Tags) > 0 {
		var tagIDs []int64
		for _, name := range in.Tags {
			tagID, err := mysqlDAO.GetOrCreateTag(name, snowflake.GenID())
			if err != nil {
				continue
			}
			tagIDs = append(tagIDs, tagID)
			_ = redisDAO.UpdateTagHotScore(ctx, name, 1)
		}
		_ = mysqlDAO.UpsertArticleTags(a.ID, tagIDs, "manual")
	}

	// 异步任务：摘要、AI 标签、向量索引（失败不影响发布）
	_ = worker.EnqueueTask(ctx, redisDAO.TaskArticleSummary, a.ID, nil)
	_ = worker.EnqueueTask(ctx, redisDAO.TaskArticleTag, a.ID, nil)
	_ = worker.EnqueueTask(ctx, redisDAO.TaskArticleIndex, a.ID, nil)

	// 更新热榜（新文章初始分 = 0）
	_ = redisDAO.UpdateHotScore(ctx, a.ID, 0)

	return a.ID, nil
}

// GetArticleDetail 获取文章详情（优先读缓存）
func GetArticleDetail(ctx context.Context, articleID int64) (*model.ArticleDetail, error) {
	// 尝试缓存
	if cached, err := redisDAO.GetArticleCache(ctx, articleID); err == nil && cached != nil {
		return cached, nil
	}

	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}
	if a == nil {
		return nil, ErrArticleNotExist
	}

	tags, _ := mysqlDAO.GetTagsByArticleID(articleID)
	a.Tags = tags

	// 异步更新浏览数和热榜（失败不影响主流程）
	go func() {
		_ = mysqlDAO.IncrViewCount(articleID)
		score := calcHotScore(a.ViewCount+1, a.LikeCount, a.FavoriteCount, a.CommentCount, a.CreatedAt)
		_ = redisDAO.UpdateHotScore(context.Background(), articleID, score)
	}()

	// 写缓存
	_ = redisDAO.SetArticleCache(ctx, a)

	return a, nil
}

// ListArticles 分页列表
func ListArticles(page, pageSize int) ([]*model.ArticleListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return mysqlDAO.ListArticles(page, pageSize)
}

// GetHotArticles 热榜前 N 篇
func GetHotArticles(ctx context.Context, topN int64) ([]*model.ArticleListItem, error) {
	ids, err := redisDAO.GetHotArticleIDs(ctx, topN)
	if err != nil || len(ids) == 0 {
		// 降级：直接查库返回最近发布
		list, _, err := mysqlDAO.ListArticles(1, int(topN))
		return list, err
	}
	return mysqlDAO.GetArticlesByIDs(ids)
}

// UpdateArticle 编辑文章
func UpdateArticle(ctx context.Context, articleID, userID int64, in *UpdateArticleInput) error {
	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil {
		return err
	}
	if a == nil {
		return ErrArticleNotExist
	}
	if a.AuthorID != userID {
		return ErrForbidden
	}
	if err := mysqlDAO.UpdateArticle(articleID, in.Title, in.Content, in.Cover); err != nil {
		return err
	}
	_ = redisDAO.DelArticleCache(ctx, articleID)

	// 编辑后重新生成摘要和向量索引
	_ = worker.EnqueueTask(ctx, redisDAO.TaskArticleSummary, articleID, nil)
	_ = worker.EnqueueTask(ctx, redisDAO.TaskArticleReindex, articleID, nil)
	return nil
}

// DeleteArticle 软删除文章
func DeleteArticle(ctx context.Context, articleID, userID int64) error {
	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil {
		return err
	}
	if a == nil {
		return ErrArticleNotExist
	}
	if a.AuthorID != userID {
		return ErrForbidden
	}
	if err := mysqlDAO.SoftDeleteArticle(articleID, userID); err != nil {
		return err
	}
	_ = redisDAO.DelArticleCache(ctx, articleID)

	// 删除文章时清理向量索引
	_ = worker.EnqueueTask(ctx, redisDAO.TaskArticleDeleteIndex, articleID, nil)
	return nil
}

// LikeArticle 点赞/取消点赞，返回当前是否已点赞
func LikeArticle(ctx context.Context, articleID, userID int64) (bool, error) {
	liked, err := mysqlDAO.ExistsUserLike(userID, articleID)
	if err != nil {
		return false, err
	}

	if liked {
		if err := mysqlDAO.DeleteUserLike(userID, articleID); err != nil {
			return false, err
		}
		_ = mysqlDAO.IncrLikeCount(articleID, -1)
		go refreshHotScore(context.Background(), articleID)
		return false, nil
	}

	if err := mysqlDAO.CreateUserLike(userID, articleID); err != nil {
		return false, err
	}
	_ = mysqlDAO.IncrLikeCount(articleID, 1)
	go refreshHotScore(context.Background(), articleID)
	return true, nil
}

// FavoriteArticle 收藏/取消收藏，返回当前是否已收藏
func FavoriteArticle(ctx context.Context, articleID, userID int64) (bool, error) {
	exists, err := mysqlDAO.ExistsFavorite(userID, articleID)
	if err != nil {
		return false, err
	}

	if exists {
		if err := mysqlDAO.DeleteFavorite(userID, articleID); err != nil {
			return false, err
		}
		_ = mysqlDAO.IncrFavoriteCount(articleID, -1)
		go refreshHotScore(context.Background(), articleID)
		return false, nil
	}

	f := &model.Favorite{UserID: userID, ArticleID: articleID}
	if err := mysqlDAO.CreateFavorite(f); err != nil {
		return false, err
	}
	_ = mysqlDAO.IncrFavoriteCount(articleID, 1)
	go refreshHotScore(context.Background(), articleID)
	return true, nil
}

// SearchResult 搜索结果，含文章列表和 AI 总结
type SearchResult struct {
	List    []*model.ArticleListItem `json:"list"`
	Total   int                      `json:"total"`
	Summary string                   `json:"summary"` // AI 搜索总结，可为空
}

// SearchArticles 关键词+语义搜索+AI总结
// 降级策略：Milvus 不可用时只返回关键词结果；AI 总结失败时返回列表不含 summary
func SearchArticles(ctx context.Context, keyword string, page, pageSize int) (*SearchResult, error) {
	// 1. MySQL 关键词搜索
	kwStart := time.Now()
	kwList, total, err := mysqlDAO.SearchArticles(keyword, page, pageSize)
	monitor.ObserveArticleSearch("keyword", time.Since(kwStart))
	if err != nil {
		return nil, err
	}

	// 2. Milvus 语义搜索（降级：失败则跳过）
	var semanticIDs []int64
	semanticStart := time.Now()
	if queryVec, embErr := aiEmbed.EmbedText(ctx, keyword); embErr == nil {
		semanticIDs, _ = milvusDAO.SearchSimilar(ctx, &settings.Conf.Milvus, queryVec, 10)
	}
	monitor.ObserveArticleSearch("semantic", time.Since(semanticStart))

	// 3. 结果融合（语义结果补充在关键词结果后，去重）
	merged := mergeResults(kwList, semanticIDs)

	// 4. AI 搜索总结（降级：超时或失败则 summary=""）
	var aiSummary string
	if len(merged) > 0 {
		summaryStart := time.Now()
		summaries := extractSummaries(merged, 5)
		aiSummary, _ = aiSkill.SearchSummarySkill(ctx, keyword, summaries)
		monitor.ObserveArticleSearch("summary", time.Since(summaryStart))
	}

	return &SearchResult{
		List:    merged,
		Total:   total,
		Summary: aiSummary,
	}, nil
}

// mergeResults 将语义搜索的文章 ID 对应的列表项合并到关键词结果后（去重）
func mergeResults(kwList []*model.ArticleListItem, semanticIDs []int64) []*model.ArticleListItem {
	if len(semanticIDs) == 0 {
		return kwList
	}
	seen := make(map[int64]struct{}, len(kwList))
	for _, a := range kwList {
		seen[a.ID] = struct{}{}
	}
	extras, _ := mysqlDAO.GetArticlesByIDs(semanticIDs)
	for _, a := range extras {
		if _, dup := seen[a.ID]; !dup {
			kwList = append(kwList, a)
			seen[a.ID] = struct{}{}
		}
	}
	return kwList
}

// extractSummaries 从文章列表提取前 n 条摘要（用于 AI 总结）
func extractSummaries(list []*model.ArticleListItem, n int) []string {
	var summaries []string
	for i, a := range list {
		if i >= n {
			break
		}
		if a.Summary != "" {
			summaries = append(summaries, a.Summary)
		} else {
			summaries = append(summaries, a.Title)
		}
	}
	return summaries
}

// calcHotScore 计算热度分：view*1 + like*5 + favorite*8 + comment*3 - age_hours*0.2
func calcHotScore(view, like, favorite, comment int, createdAt time.Time) float64 {
	ageHours := time.Since(createdAt).Hours()
	score := float64(view)*1 + float64(like)*5 + float64(favorite)*8 + float64(comment)*3 - ageHours*0.2
	return math.Max(score, 0)
}

// refreshHotScore 从 DB 读取文章最新统计值，重新计算并更新 Redis 热榜分数
// 用于点赞/收藏/评论后异步调用，失败不影响主流程
func refreshHotScore(ctx context.Context, articleID int64) {
	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil || a == nil {
		return
	}
	score := calcHotScore(a.ViewCount, a.LikeCount, a.FavoriteCount, a.CommentCount, a.CreatedAt)
	_ = redisDAO.UpdateHotScore(ctx, articleID, score)
}

func ListUserFavorites(userID int64, page, pageSize int) ([]*model.ArticleListItem, int, error) {
	return mysqlDAO.ListFavoritesByUserID(userID, page, pageSize)
}

func ListUserLikes(userID int64, page, pageSize int) ([]*model.ArticleListItem, int, error) {
	return mysqlDAO.ListUserLikesByUserID(userID, page, pageSize)
}
