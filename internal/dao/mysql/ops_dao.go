package mysql

import (
	"context"
	"errors"
	"time"

	"techmind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ── OpsReport ───────────────────────────────────────────────

// CreateOpsReport 写入诊断报告
func CreateOpsReport(r *model.OpsReport) error {
	return DB.Create(r).Error
}

// PrepareOpsReport 按 task_key 幂等创建或复用报告。
// Worker 重试和 stale claim 不会再为同一任务生成多份报告。
func PrepareOpsReport(candidate *model.OpsReport) (report *model.OpsReport, completed bool, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		var existing model.OpsReport
		queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("task_key = ?", candidate.TaskKey).First(&existing).Error
		if queryErr == nil {
			if existing.Status == "done" {
				report = &existing
				completed = true
				return nil
			}
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"alert_id": candidate.AlertID, "incident_id": candidate.IncidentID,
				"trigger_type": candidate.TriggerType, "status": "running",
			}).Error; err != nil {
				return err
			}
			existing.AlertID = candidate.AlertID
			existing.IncidentID = candidate.IncidentID
			existing.TriggerType = candidate.TriggerType
			existing.Status = "running"
			report = &existing
			return nil
		}
		if !errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return queryErr
		}
		if err := tx.Create(candidate).Error; err != nil {
			return err
		}
		report = candidate
		return nil
	})
	return report, completed, err
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

// CreateOpsToolCall 写入一次 Agent 工具调用审计，失败不应中断诊断主流程。
func CreateOpsToolCall(ctx context.Context, call *model.OpsToolCall) error {
	return DB.WithContext(ctx).Create(call).Error
}

// ListOpsToolCalls 按调用顺序读取报告的证据链。
func ListOpsToolCalls(reportID int64) ([]*model.OpsToolCall, error) {
	var calls []*model.OpsToolCall
	err := DB.Where("report_id = ?", reportID).Order("id ASC").Find(&calls).Error
	return calls, err
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

// GetChangesBetween 查询精确诊断时间窗内的部署变更。
func GetChangesBetween(ctx context.Context, service string, start, end time.Time) ([]*model.DeploymentChange, error) {
	q := DB.WithContext(ctx).Model(&model.DeploymentChange{}).
		Where("changed_at BETWEEN ? AND ?", start, end)
	if service != "" {
		q = q.Where("service = ?", service)
	}
	var list []*model.DeploymentChange
	err := q.Order("changed_at DESC").Find(&list).Error
	return list, err
}
