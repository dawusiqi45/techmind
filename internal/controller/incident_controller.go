package controller

import (
	"strconv"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListIncidents GET /api/v1/admin/incidents
func ListIncidents(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := mysqlDAO.ListIncidents(status, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

// GetIncident GET /api/v1/admin/incidents/:id
func GetIncident(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	incident, err := mysqlDAO.GetIncidentByID(id)
	if err != nil || incident == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}
	alerts, err := mysqlDAO.GetAlertsByIncidentID(id)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"incident": incident, "alerts": alerts})
}

// ResolveIncident POST /api/v1/admin/incidents/:id/resolve
// 故障关闭由管理员确认；此操作不会改变告警本身的 firing/resolved 状态。
func ResolveIncident(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	incident, err := mysqlDAO.GetIncidentByID(id)
	if err != nil || incident == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}
	if err := mysqlDAO.ResolveIncident(id); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, nil)
}
