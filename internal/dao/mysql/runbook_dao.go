package mysql

import (
	"errors"

	"techmind/internal/model"

	"gorm.io/gorm"
)

// CreateRunbook 写入 Runbook
func CreateRunbook(r *model.Runbook) error {
	return DB.Create(r).Error
}

// GetRunbookByID 按 ID 查询
func GetRunbookByID(id int64) (*model.Runbook, error) {
	var r model.Runbook
	err := DB.First(&r, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &r, err
}

// ListRunbooks 分页查询（按更新时间倒序）
func ListRunbooks(alertName string, page, pageSize int) ([]*model.Runbook, int, error) {
	offset := (page - 1) * pageSize
	q := DB.Model(&model.Runbook{})
	if alertName != "" {
		q = q.Where("alert_name = ?", alertName)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.Runbook
	err := q.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// GetUnindexedRunbooks 查询未向量化的 Runbook
func GetUnindexedRunbooks(limit int) ([]*model.Runbook, error) {
	var list []*model.Runbook
	err := DB.Where("index_status = 0").Limit(limit).Find(&list).Error
	return list, err
}

// UpdateRunbookIndexStatus 更新 Runbook 向量索引状态
func UpdateRunbookIndexStatus(id int64, status int8) error {
	return DB.Model(&model.Runbook{}).Where("id = ?", id).
		Update("index_status", status).Error
}

// SearchRunbooksByAlertName 按告警名称查找相关 Runbook
func SearchRunbooksByAlertName(alertName, service string, limit int) ([]*model.Runbook, error) {
	q := DB.Model(&model.Runbook{})
	if alertName != "" {
		q = q.Where("alert_name = ? OR alert_name = ''", alertName)
	}
	if service != "" {
		q = q.Where("service = ? OR service = ''", service)
	}
	var list []*model.Runbook
	err := q.Order("updated_at DESC").Limit(limit).Find(&list).Error
	return list, err
}
