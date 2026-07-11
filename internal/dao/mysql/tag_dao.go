package mysql

import (
	"errors"

	"techmind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetOrCreateTag 获取或创建标签，返回 tag ID
func GetOrCreateTag(name string, id int64) (int64, error) {
	var tag model.Tag
	err := DB.Where("name = ?", name).First(&tag).Error
	if err == nil {
		return tag.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	tag = model.Tag{ID: id, Name: name}
	if err := DB.Create(&tag).Error; err != nil {
		return 0, err
	}
	return tag.ID, nil
}

// GetTagsByArticleID 查询文章关联标签名列表
func GetTagsByArticleID(articleID int64) ([]string, error) {
	var names []string
	err := DB.Raw(`
		SELECT t.name FROM tag t
		JOIN article_tag at ON at.tag_id = t.id
		WHERE at.article_id = ?`, articleID).Scan(&names).Error
	return names, err
}

// UpsertArticleTags 批量绑定文章标签，已存在则忽略
func UpsertArticleTags(articleID int64, tagIDs []int64, source string) error {
	if len(tagIDs) == 0 {
		return nil
	}
	tags := make([]model.ArticleTag, 0, len(tagIDs))
	for _, tid := range tagIDs {
		tags = append(tags, model.ArticleTag{
			ArticleID: articleID,
			TagID:     tid,
			Source:    source,
		})
	}
	return DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&tags).Error
}

// ReplaceArticleTags 用给定集合替换文章某一来源的标签，其他来源（如 AI）保持不变。
func ReplaceArticleTags(articleID int64, tagIDs []int64, source string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return replaceArticleTags(tx, articleID, tagIDs, source)
	})
}

func replaceArticleTags(tx *gorm.DB, articleID int64, tagIDs []int64, source string) error {
	if err := tx.Where("article_id = ? AND source = ?", articleID, source).
		Delete(&model.ArticleTag{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	tags := make([]model.ArticleTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		tags = append(tags, model.ArticleTag{ArticleID: articleID, TagID: tagID, Source: source})
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&tags).Error
}

// ListHotTags 按热度分倒序返回 topN 标签
func ListHotTags(topN int) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := DB.Order("hot_score DESC").Limit(topN).Find(&tags).Error
	return tags, err
}

// ListAllTags 全量标签
func ListAllTags() ([]*model.Tag, error) {
	var tags []*model.Tag
	err := DB.Order("hot_score DESC").Find(&tags).Error
	return tags, err
}
