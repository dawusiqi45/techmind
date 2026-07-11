package mysql

import (
	"errors"

	"techmind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateComment 插入评论
func CreateComment(c *model.Comment) error {
	return DB.Create(c).Error
}

// CreateCommentWithCount 在同一事务内创建评论并更新文章评论数。
func CreateCommentWithCount(c *model.Comment) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = 1", c.ArticleID).First(&article).Error; err != nil {
			return err
		}
		if err := tx.Create(c).Error; err != nil {
			return err
		}
		return tx.Model(&model.Article{}).Where("id = ?", c.ArticleID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

// GetCommentByID 按 ID 查询，未找到返回 nil
func GetCommentByID(id int64) (*model.Comment, error) {
	var c model.Comment
	err := DB.Where("id = ? AND status = 1", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

// ListCommentsByArticle 查询文章的一级评论（含作者信息）
func ListCommentsByArticle(articleID int64) ([]*model.CommentDetail, error) {
	var list []*model.CommentDetail
	err := DB.Raw(`
		SELECT c.*, u.username AS author_name, u.avatar AS author_avatar
		FROM comment c
		JOIN user u ON u.id = c.author_id
		WHERE c.article_id = ? AND c.parent_id = 0 AND c.status = 1
		ORDER BY c.created_at ASC`, articleID).Scan(&list).Error
	return list, err
}

// ListRepliesByParent 查询某评论的回复列表
func ListRepliesByParent(parentID int64) ([]*model.CommentDetail, error) {
	var list []*model.CommentDetail
	err := DB.Raw(`
		SELECT c.*, u.username AS author_name, u.avatar AS author_avatar
		FROM comment c
		JOIN user u ON u.id = c.author_id
		WHERE c.parent_id = ? AND c.status = 1
		ORDER BY c.created_at ASC`, parentID).Scan(&list).Error
	return list, err
}

// SoftDeleteComment 软删除评论（仅允许作者操作）
func SoftDeleteComment(id, authorID int64) error {
	return DB.Model(&model.Comment{}).
		Where("id = ? AND author_id = ?", id, authorID).
		Update("status", -1).Error
}
