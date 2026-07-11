package logic

import (
	"context"
	"errors"
	"fmt"

	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/model"
	"techmind/internal/pkg/snowflake"
)

var ErrCommentNotExist = errors.New("comment not found")

// CreateCommentInput 发布评论参数
type CreateCommentInput struct {
	ArticleID int64
	ParentID  int64 // 0=一级评论
	Content   string
}

// CreateComment 发布评论
func CreateComment(authorID int64, in *CreateCommentInput) (int64, error) {
	// 若是回复，验证父评论存在
	if in.ParentID != 0 {
		parent, err := mysqlDAO.GetCommentByID(in.ParentID)
		if err != nil {
			return 0, fmt.Errorf("create comment: check parent: %w", err)
		}
		if parent == nil {
			return 0, ErrCommentNotExist
		}
		if parent.ArticleID != in.ArticleID {
			return 0, ErrCommentNotExist
		}
	}

	c := &model.Comment{
		ID:        snowflake.GenID(),
		ArticleID: in.ArticleID,
		AuthorID:  authorID,
		ParentID:  in.ParentID,
		Content:   in.Content,
		Status:    1,
	}
	if err := mysqlDAO.CreateCommentWithCount(c); err != nil {
		return 0, err
	}
	_ = redisDAO.DelArticleCache(context.Background(), in.ArticleID)
	go refreshHotScore(context.Background(), in.ArticleID)
	return c.ID, nil
}

// ListComments 获取文章评论树（一级 + 回复）
func ListComments(articleID int64) ([]*model.CommentDetail, error) {
	roots, err := mysqlDAO.ListCommentsByArticle(articleID)
	if err != nil {
		return nil, err
	}

	for _, root := range roots {
		replies, err := mysqlDAO.ListRepliesByParent(root.ID)
		if err != nil {
			continue
		}
		root.Replies = replies
	}
	return roots, nil
}

// GetListTags 获取全量标签列表
func GetListTags() ([]*model.Tag, error) {
	return mysqlDAO.ListAllTags()
}
