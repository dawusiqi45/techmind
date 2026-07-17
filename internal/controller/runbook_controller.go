package controller

import (
	"strconv"

	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/model"
	"techmind/internal/pkg/response"
	"techmind/internal/pkg/snowflake"
	"techmind/internal/worker"

	"github.com/gin-gonic/gin"
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

	// 交给 Redis Stream Worker 执行，复用既有重试、死信和崩溃接管能力。
	if err := worker.EnqueueTask(c.Request.Context(), redisDAO.TaskRunbookIndex, rb.ID, nil); err != nil {
		_ = mysqlDAO.UpdateRunbookIndexStatus(rb.ID, -1)
		response.Fail(c, response.CodeServerError)
		return
	}

	response.OK(c, gin.H{"id": strconv.FormatInt(rb.ID, 10), "index_status": 0})
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
