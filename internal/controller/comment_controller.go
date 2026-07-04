package controller

import (
	"errors"

	"techmind/internal/logic"
	"techmind/internal/middleware"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateCommentReq 发布评论请求体
type CreateCommentReq struct {
	ParentID int64  `json:"parent_id"`
	Content  string `json:"content" binding:"required,min=1,max=2000"`
}

// CreateComment POST /api/v1/articles/:id/comments
func CreateComment(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	articleID, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	var req CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	commentID, err := logic.CreateComment(uid, &logic.CreateCommentInput{
		ArticleID: articleID,
		ParentID:  req.ParentID,
		Content:   req.Content,
	})
	if err != nil {
		if errors.Is(err, logic.ErrCommentNotExist) {
			response.FailWithMsg(c, response.CodeInvalidParam, "parent comment not found")
			return
		}
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"id": commentID})
}

// ListComments GET /api/v1/articles/:id/comments
func ListComments(c *gin.Context) {
	articleID, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	list, err := logic.ListComments(articleID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, list)
}

// ListTags GET /api/v1/tags
func ListTags(c *gin.Context) {
	tags, err := logic.GetListTags()
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, tags)
}
