package mysql

import (
	"errors"
	"time"

	"techmind/internal/model"

	"gorm.io/gorm"
)

// ── OpsReport ───────────────────────────────────────────────

// CreateOpsReport 写入诊断报告
func CreateOpsReport(r *model.OpsReport) error {
	return DB.Create(r).Error
}

// GetOpsReportByID 按 ID 查询
func GetOpsReportByID(id int64) (*model.OpsReport, error) {
	var r model.OpsReport
	err := DB.First(&r, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &r, err
}

// ListOpsReports 分页查询诊断报告（按创建时间倒序）
func ListOpsReports(page, pageSize int) ([]*model.OpsReport, int, error) {
	offset := (page - 1) * pageSize
	var total int64
	if err := DB.Model(&model.OpsReport{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.OpsReport
	err := DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// UpdateOpsReportStatus 更新报告状态（running/done/failed）
func UpdateOpsReportStatus(id int64, status string) error {
	return DB.Model(&model.OpsReport{}).Where("id = ?", id).Update("status", status).Error
}

// ── DeploymentChange ────────────────────────────────────────

// CreateDeploymentChange 写入部署变更
func CreateDeploymentChange(c *model.DeploymentChange) error {
	return DB.Create(c).Error
}

// ListDeploymentChanges 分页查询部署变更（按变更时间倒序）
func ListDeploymentChanges(service string, page, pageSize int) ([]*model.DeploymentChange, int, error) {
	offset := (page - 1) * pageSize
	q := DB.Model(&model.DeploymentChange{})
	if service != "" {
		q = q.Where("service = ?", service)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.DeploymentChange
	err := q.Order("changed_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// GetRecentChanges 查询告警发生时间前后 window 内的部署变更（用于变更关联）
func GetRecentChanges(service string, alertTime time.Time, windowMinutes int) ([]*model.DeploymentChange, error) {
	start := alertTime.Add(-time.Duration(windowMinutes) * time.Minute)
	end := alertTime.Add(time.Duration(windowMinutes) * time.Minute)

	q := DB.Model(&model.DeploymentChange{}).
		Where("changed_at BETWEEN ? AND ?", start, end)
	if service != "" {
		q = q.Where("service = ?", service)
	}

	var list []*model.DeploymentChange
	err := q.Order("changed_at DESC").Find(&list).Error
	return list, err
}
