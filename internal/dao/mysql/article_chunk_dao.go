package mysql

import (
	"techmind/internal/model"
)

// BatchCreateArticleChunks 批量写入文章 chunk 正文
func BatchCreateArticleChunks(chunks []*model.ArticleChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return DB.CreateInBatches(chunks, 50).Error
}

// DeleteArticleChunks 按 article_id 删除所有 chunk（文章删除/reindex 时调用）
func DeleteArticleChunks(articleID int64) error {
	return DB.Where("article_id = ?", articleID).Delete(&model.ArticleChunk{}).Error
}
