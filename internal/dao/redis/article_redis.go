package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"techmind/internal/model"
	"techmind/internal/monitor"

	goredis "github.com/redis/go-redis/v9"
)

// goredisZ 构造 ZSet member
func goredisZ(score float64, articleID int64) goredis.Z {
	return goredis.Z{Score: score, Member: strconv.FormatInt(articleID, 10)}
}

const (
	keyArticleDetail = "tm:article:detail:%d" // String: 文章详情缓存
	keyArticleHot    = "tm:article:hot"       // ZSet: 全站热榜
	keyTagHot        = "tm:tag:hot"           // ZSet: 热门标签
	keyArticleLiked  = "tm:article:liked:%d"  // Set: 点赞防重复
	articleCacheTTL  = 10 * time.Minute
)

// ── 文章详情缓存 ────────────────────────────────────────────

// SetArticleCache 缓存文章详情 JSON
func SetArticleCache(ctx context.Context, a *model.ArticleDetail) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(keyArticleDetail, a.ID)
	return RDB.Set(ctx, key, data, articleCacheTTL).Err()
}

// GetArticleCache 读取文章详情缓存，未命中返回 nil
func GetArticleCache(ctx context.Context, articleID int64) (*model.ArticleDetail, error) {
	key := fmt.Sprintf(keyArticleDetail, articleID)
	data, err := RDB.Get(ctx, key).Bytes()
	if err != nil {
		monitor.IncCacheMiss()
		return nil, nil // 未命中视为缓存 miss
	}
	a := &model.ArticleDetail{}
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	monitor.IncCacheHit()
	return a, nil
}

// DelArticleCache 删除文章缓存（编辑/删除时调用）
func DelArticleCache(ctx context.Context, articleID int64) error {
	key := fmt.Sprintf(keyArticleDetail, articleID)
	return RDB.Del(ctx, key).Err()
}

// ── 热榜 ZSet ───────────────────────────────────────────────

// UpdateHotScore 更新热榜分数
// score = view*1 + like*5 + favorite*8 + comment*3 - age_hours*0.2
func UpdateHotScore(ctx context.Context, articleID int64, score float64) error {
	return RDB.ZAdd(ctx, keyArticleHot, goredisZ(score, articleID)).Err()
}

// GetHotArticleIDs 获取热榜 topN 文章 ID（分数降序）
func GetHotArticleIDs(ctx context.Context, topN int64) ([]int64, error) {
	vals, err := RDB.ZRevRange(ctx, keyArticleHot, 0, topN-1).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(vals))
	for _, v := range vals {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ── 点赞 Set（防重复）───────────────────────────────────────

// LikeArticle 点赞：将 userID 加入点赞 Set，返回是否为新增（false=已点赞）
func LikeArticle(ctx context.Context, articleID, userID int64) (bool, error) {
	key := fmt.Sprintf(keyArticleLiked, articleID)
	added, err := RDB.SAdd(ctx, key, userID).Result()
	return added > 0, err
}

// UnlikeArticle 取消点赞
func UnlikeArticle(ctx context.Context, articleID, userID int64) (bool, error) {
	key := fmt.Sprintf(keyArticleLiked, articleID)
	removed, err := RDB.SRem(ctx, key, userID).Result()
	return removed > 0, err
}

// IsLiked 判断用户是否已点赞
func IsLiked(ctx context.Context, articleID, userID int64) (bool, error) {
	key := fmt.Sprintf(keyArticleLiked, articleID)
	return RDB.SIsMember(ctx, key, userID).Result()
}

// ── 热门标签 ZSet ────────────────────────────────────────────

// UpdateTagHotScore 更新热门标签分数
func UpdateTagHotScore(ctx context.Context, tagName string, delta float64) error {
	return RDB.ZIncrBy(ctx, keyTagHot, delta, tagName).Err()
}

// GetHotTagNames 获取热门标签名列表 topN
func GetHotTagNames(ctx context.Context, topN int64) ([]string, error) {
	return RDB.ZRevRange(ctx, keyTagHot, 0, topN-1).Result()
}
