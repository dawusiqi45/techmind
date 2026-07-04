package controller

import (
	"strconv"

	"techmind/internal/logic"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AlertWebhook POST /api/v1/admin/alerts/webhook
// 接收 Alertmanager 推送
func AlertWebhook(c *gin.Context) {
	var payload logic.AlertmanagerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}
	if err := logic.ReceiveAlertWebhook(&payload); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, nil)
}

// ListAlerts GET /api/v1/admin/alerts
func ListAlerts(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := logic.ListAlerts(status, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

// GetAlertDetail GET /api/v1/admin/alerts/:id
func GetAlertDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	detail, err := logic.GetAlertDetail(id)
	if err != nil || detail == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}
	response.OK(c, detail)
}

// AcknowledgeAlert POST /api/v1/admin/alerts/:id/ack
func AcknowledgeAlert(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	if err := logic.AcknowledgeAlert(id); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, nil)
}
