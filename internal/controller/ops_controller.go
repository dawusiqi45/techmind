package controller

import (
	"strconv"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/pkg/response"
	"techmind/internal/worker"

	"github.com/gin-gonic/gin"
)

// ManualDiagnose POST /api/v1/admin/ops/diagnose
// 手动触发 SRE 诊断（写入 ops_tasks Stream，由 OpsWorker 异步消费）
func ManualDiagnose(c *gin.Context) {
	var req struct {
		Service   string `json:"service"`
		AlertName string `json:"alert_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := worker.EnqueueDiagnoseTask(c.Request.Context(), 0, "manual", req.Service, req.AlertName); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"message": "诊断任务已入队，请稍后查询报告列表"})
}

// AlertDiagnose POST /api/v1/admin/alerts/:id/diagnose
// 对指定告警触发 SRE 诊断（入队异步执行）
func AlertDiagnose(c *gin.Context) {
	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	event, err := mysqlDAO.GetAlertEventByID(alertID)
	if err != nil || event == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}

	if err := worker.EnqueueDiagnoseTask(c.Request.Context(), alertID, "alert", event.Service, event.AlertName); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"message": "诊断任务已入队，请稍后查询报告列表"})
}

// ListOpsReports GET /api/v1/admin/ops/reports
func ListOpsReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := mysqlDAO.ListOpsReports(page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

// GetOpsReport GET /api/v1/admin/ops/reports/:id
func GetOpsReport(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	report, err := mysqlDAO.GetOpsReportByID(id)
	if err != nil || report == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}
	response.OK(c, report)
}
