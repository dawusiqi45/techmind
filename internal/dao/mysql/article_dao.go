package mysql

import (
	"techmind/internal/model"

	"gorm.io/gorm"
)

const articleListSelect = `
	a.id, a.author_id, u.username AS author_name,
	a.title, a.summary, a.cover,
	a.view_count, a.like_count, a.favorite_count, a.comment_count, a.created_at`

// CreateArticle 插入文章
func CreateArticle(a *model.Article) error {
	return DB.Create(a).Error
}

// CreateArticleWithTags 在同一事务内创建文章及其手动标签。
func CreateArticleWithTags(a *model.Article, tagIDs []int64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
		return replaceArticleTags(tx, a.ID, tagIDs, "manual")
	})
}

// GetArticleByID 按 ID 查询文章（含作者名），未找到返回 nil
func GetArticleByID(id int64) (*model.ArticleDetail, error) {
	var a model.ArticleDetail
	err := DB.Raw(`
		SELECT a.*, u.username AS author_name
		FROM article a
		JOIN user u ON u.id = a.author_id
		WHERE a.id = ? AND a.status != -1 LIMIT 1`, id).Scan(&a).Error
	if err != nil {
		return nil, err
	}
	if a.ID == 0 {
		return nil, nil
	}
	return &a, nil
}

// ListArticles 分页查询文章列表（按创建时间倒序）
func ListArticles(page, pageSize int) ([]*model.ArticleListItem, int, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := DB.Model(&model.Article{}).Where("status = 1").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.ArticleListItem
	err := DB.Raw(`
		SELECT `+articleListSelect+`
		FROM article a
		JOIN user u ON u.id = a.author_id
		WHERE a.status = 1
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?`, pageSize, offset).Scan(&list).Error
	return list, int(total), err
}

// ListArticlesByTag 按标签分页查询
func ListArticlesByTag(tagID int64, page, pageSize int) ([]*model.ArticleListItem, int, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := DB.Raw(`
		SELECT COUNT(1) FROM article a
		JOIN article_tag at ON at.article_id = a.id
		WHERE at.tag_id = ? AND a.status = 1`, tagID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.ArticleListItem
	err := DB.Raw(`
		SELECT `+articleListSelect+`
		FROM article a
		JOIN user u ON u.id = a.author_id
		JOIN article_tag at ON at.article_id = a.id
		WHERE at.tag_id = ? AND a.status = 1
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?`, tagID, pageSize, offset).Scan(&list).Error
	return list, int(total), err
}

// SearchArticles MySQL 关键词搜索（标题+正文 LIKE）
func SearchArticles(keyword string, page, pageSize int) ([]*model.ArticleListItem, int, error) {
	offset := (page - 1) * pageSize
	like := "%" + keyword + "%"

	var total int64
	if err := DB.Model(&model.Article{}).
		Where("status = 1 AND (title LIKE ? OR content LIKE ?)", like, like).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.ArticleListItem
	err := DB.Raw(`
		SELECT `+articleListSelect+`
		FROM article a
		JOIN user u ON u.id = a.author_id
		WHERE a.status = 1 AND (a.title LIKE ? OR a.content LIKE ?)
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?`, like, like, pageSize, offset).Scan(&list).Error
	return list, int(total), err
}

// UpdateArticle 更新标题、正文和封面（编辑场景），重置 index_status=0
func UpdateArticle(id int64, title, content, cover string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":        title,
			"content":      content,
			"cover":        cover,
			"index_status": 0,
		}).Error
}

// UpdateArticleWithTags 在同一事务内更新文章正文和手动标签。
func UpdateArticleWithTags(id int64, title, content, cover string, tagIDs []int64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Article{}).Where("id = ?", id).
			Updates(map[string]interface{}{
				"title": title, "content": content, "cover": cover, "index_status": 0,
			}).Error; err != nil {
			return err
		}
		return replaceArticleTags(tx, id, tagIDs, "manual")
	})
}

// SoftDeleteArticle 软删除（status = -1），仅允许作者操作
func SoftDeleteArticle(id, authorID int64) error {
	return DB.Model(&model.Article{}).
		Where("id = ? AND author_id = ?", id, authorID).
		Update("status", -1).Error
}

// IncrViewCount 文章浏览数 +1
func IncrViewCount(id int64) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrLikeCount 文章点赞数增减（delta = +1 或 -1）
func IncrLikeCount(id int64, delta int) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
}

// IncrFavoriteCount 收藏数增减
func IncrFavoriteCount(id int64, delta int) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", delta)).Error
}

// IncrCommentCount 评论数增减
func IncrCommentCount(id int64, delta int) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

// UpdateArticleSummary 更新 AI 生成摘要
func UpdateArticleSummary(id int64, summary string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).
		Update("summary", summary).Error
}

// GetArticlesByIDs 按 ID 列表批量查询（热榜场景）
func GetArticlesByIDs(ids []int64) ([]*model.ArticleListItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*model.ArticleListItem
	err := DB.Raw(`
		SELECT `+articleListSelect+`
		FROM article a
		JOIN user u ON u.id = a.author_id
		WHERE a.id IN ? AND a.status = 1`, ids).Scan(&list).Error
	return list, err
}
