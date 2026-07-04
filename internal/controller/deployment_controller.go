package controller

import (
	"strconv"
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/pkg/response"
	"techmind/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
)

// RecordDeploymentChange POST /api/v1/admin/deployment-changes
func RecordDeploymentChange(c *gin.Context) {
	var req struct {
		Service   string `json:"service"    binding:"required"`
		Namespace string `json:"namespace"`
		Image     string `json:"image"`
		OldImage  string `json:"old_image"`
		Replicas  int    `json:"replicas"`
		ChangedBy string `json:"changed_by"`
		Source    string `json:"source"` // helm/kubectl/argocd/manual
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Source == "" {
		req.Source = "manual"
	}

	change := &model.DeploymentChange{
		ID:        snowflake.GenID(),
		Service:   req.Service,
		Namespace: req.Namespace,
		Image:     req.Image,
		OldImage:  req.OldImage,
		Replicas:  req.Replicas,
		ChangedBy: req.ChangedBy,
		Source:    req.Source,
		ChangedAt: time.Now(),
	}
	if err := mysqlDAO.CreateDeploymentChange(change); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"id": strconv.FormatInt(change.ID, 10)})
}

// ListDeploymentChanges GET /api/v1/admin/deployment-changes
func ListDeploymentChanges(c *gin.Context) {
	service := c.DefaultQuery("service", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := mysqlDAO.ListDeploymentChanges(service, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}
