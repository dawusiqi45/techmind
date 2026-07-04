package controller

import (
	"strconv"

	"techmind/internal/agent"
	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/pkg/response"
	"techmind/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateRunbook POST /api/v1/admin/runbooks
func CreateRunbook(c *gin.Context) {
	var req struct {
		Title     string `json:"title"     binding:"required,min=1,max=200"`
		Content   string `json:"content"   binding:"required,min=1"`
		AlertName string `json:"alert_name"`
		Service   string `json:"service"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	rb := &model.Runbook{
		ID:        snowflake.GenID(),
		Title:     req.Title,
		Content:   req.Content,
		AlertName: req.AlertName,
		Service:   req.Service,
	}
	if err := mysqlDAO.CreateRunbook(rb); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}

	// 异步触发 Milvus 向量索引
	go func() {
		if err := agent.IndexRunbook(c.Request.Context(), rb.ID, rb.Title, rb.Content); err != nil {
			zap.L().Warn("index runbook failed", zap.Int64("id", rb.ID), zap.Error(err))
			_ = mysqlDAO.UpdateRunbookIndexStatus(rb.ID, -1)
			return
		}
		_ = mysqlDAO.UpdateRunbookIndexStatus(rb.ID, 1)
	}()

	response.OK(c, gin.H{"id": strconv.FormatInt(rb.ID, 10)})
}

// ListRunbooks GET /api/v1/admin/runbooks
func ListRunbooks(c *gin.Context) {
	alertName := c.DefaultQuery("alert_name", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := mysqlDAO.ListRunbooks(alertName, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}
