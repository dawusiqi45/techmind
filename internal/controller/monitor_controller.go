package controller

import (
	"strconv"

	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// MonitorOverview GET /api/v1/admin/monitor/overview
// 返回慢请求总数、错误事件总数、AI任务 pending 数、Stream 队列长度
func MonitorOverview(c *gin.Context) {
	_, slowTotal, _ := mysqlDAO.ListSlowRequests(1, 1)
	_, errTotal, _ := mysqlDAO.ListErrorEvents("", 1, 1)

	pendingCount, _ := redisDAO.PendingAITasks(c.Request.Context())
	streamLen, _ := redisDAO.StreamLen(c.Request.Context(), redisDAO.StreamAITasks)

	response.OK(c, gin.H{
		"slow_request_total":  slowTotal,
		"error_event_total":   errTotal,
		"ai_task_pending":     pendingCount,
		"ai_stream_length":    streamLen,
	})
}

// MonitorSlowRequests GET /api/v1/admin/monitor/slow-requests
func MonitorSlowRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := mysqlDAO.ListSlowRequests(page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

// MonitorErrors GET /api/v1/admin/monitor/errors
func MonitorErrors(c *gin.Context) {
	source := c.DefaultQuery("source", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := mysqlDAO.ListErrorEvents(source, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

// MonitorAICalls GET /api/v1/admin/monitor/ai-calls
// 返回最近 AI 调用记录
func MonitorAICalls(c *gin.Context) {
	skill := c.DefaultQuery("skill", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := mysqlDAO.ListAICallRecords(skill, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}
// 返回 AI 任务 Stream 的 pending 数和队列长度
func MonitorQueues(c *gin.Context) {
	ctx := c.Request.Context()
	pending, _ := redisDAO.PendingAITasks(ctx)
	length, _ := redisDAO.StreamLen(ctx, redisDAO.StreamAITasks)
	deadLen, _ := redisDAO.StreamLen(ctx, redisDAO.StreamAIDeadLetter)

	response.OK(c, gin.H{
		"ai_task_stream":      redisDAO.StreamAITasks,
		"ai_task_pending":     pending,
		"ai_task_length":      length,
		"ai_dead_letter_len":  deadLen,
	})
}
