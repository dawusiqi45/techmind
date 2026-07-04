package controller

import (
	"errors"
	"strconv"

	"techmind/internal/logic"
	"techmind/internal/middleware"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateArticleReq 发布文章请求体
type CreateArticleReq struct {
	Title   string   `json:"title"   binding:"required,min=1,max=200"`
	Content string   `json:"content" binding:"required,min=1"`
	Cover   string   `json:"cover"   binding:"omitempty,url"`
	Tags    []string `json:"tags"`
}

// UpdateArticleReq 编辑文章请求体
type UpdateArticleReq struct {
	Title   string `json:"title"   binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
	Cover   string `json:"cover"   binding:"omitempty"`
}

// CreateArticle POST /api/v1/articles
func CreateArticle(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	var req CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	articleID, err := logic.CreateArticle(c.Request.Context(), uid, &logic.CreateArticleInput{
		Title:   req.Title,
		Content: req.Content,
		Cover:   req.Cover,
		Tags:    req.Tags,
	})
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"id": strconv.FormatInt(articleID, 10)})
}

// GetArticle GET /api/v1/articles/:id
func GetArticle(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	detail, err := logic.GetArticleDetail(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, logic.ErrArticleNotExist) {
			response.Fail(c, response.CodeArticleNotExist)
			return
		}
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, detail)
}

// ListArticles GET /api/v1/articles
func ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := logic.ListArticles(page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

// GetHotArticles GET /api/v1/articles/hot
func GetHotArticles(c *gin.Context) {
	list, err := logic.GetHotArticles(c.Request.Context(), 20)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, list)
}

// UpdateArticle PUT /api/v1/articles/:id
func UpdateArticle(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	var req UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	err = logic.UpdateArticle(c.Request.Context(), id, uid, &logic.UpdateArticleInput{
		Title:   req.Title,
		Content: req.Content,
		Cover:   req.Cover,
	})
	if err != nil {
		if errors.Is(err, logic.ErrArticleNotExist) {
			response.Fail(c, response.CodeArticleNotExist)
			return
		}
		if errors.Is(err, logic.ErrForbidden) {
			response.Fail(c, response.CodeForbidden)
			return
		}
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, nil)
}

// DeleteArticle DELETE /api/v1/articles/:id
func DeleteArticle(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	err = logic.DeleteArticle(c.Request.Context(), id, uid)
	if err != nil {
		if errors.Is(err, logic.ErrArticleNotExist) {
			response.Fail(c, response.CodeArticleNotExist)
			return
		}
		if errors.Is(err, logic.ErrForbidden) {
			response.Fail(c, response.CodeForbidden)
			return
		}
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, nil)
}

// LikeArticle POST /api/v1/articles/:id/like
func LikeArticle(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	liked, err := logic.LikeArticle(c.Request.Context(), id, uid)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"liked": liked})
}

// FavoriteArticle POST /api/v1/articles/:id/favorite
func FavoriteArticle(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	favorited, err := logic.FavoriteArticle(c.Request.Context(), id, uid)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"favorited": favorited})
}

// SearchArticles GET /api/v1/search
func SearchArticles(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		response.FailWithMsg(c, response.CodeInvalidParam, "q is required")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := logic.SearchArticles(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, result)
}

// parseID 从路径参数解析 int64 ID
func parseID(c *gin.Context, key string) (int64, error) {
	return strconv.ParseInt(c.Param(key), 10, 64)
}
