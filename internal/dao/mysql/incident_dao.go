package mysql

import (
	"errors"

	"techmind/internal/model"
	"techmind/internal/pkg/snowflake"

	"gorm.io/gorm"
)
// CreateIncident 创建故障事件并关联一组告警
func CreateIncident(title, severity string, alertIDs []int64) (*model.Incident, error) {
	incident := &model.Incident{
		ID:       snowflake.GenID(),
		Title:    title,
		Status:   "open",
		Severity: severity,
	}
	if err := DB.Create(incident).Error; err != nil {
		return nil, err
	}
	for _, aid := range alertIDs {
		link := &model.IncidentAlert{IncidentID: incident.ID, AlertID: aid}
		_ = DB.Create(link).Error
	}
	return incident, nil
}

// GetIncidentByID 按 ID 查询故障事件
func GetIncidentByID(id int64) (*model.Incident, error) {
	var inc model.Incident
	err := DB.First(&inc, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inc, err
}

// ListIncidents 分页查询故障事件（按创建时间倒序）
func ListIncidents(status string, page, pageSize int) ([]*model.Incident, int, error) {
	offset := (page - 1) * pageSize
	q := DB.Model(&model.Incident{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.Incident
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// GetAlertsByIncidentID 查询故障关联的所有告警
func GetAlertsByIncidentID(incidentID int64) ([]*model.AlertEvent, error) {
	var links []model.IncidentAlert
	if err := DB.Where("incident_id = ?", incidentID).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return nil, nil
	}
	alertIDs := make([]int64, len(links))
	for i, l := range links {
		alertIDs[i] = l.AlertID
	}
	var alerts []*model.AlertEvent
	err := DB.Where("id IN ?", alertIDs).Find(&alerts).Error
	return alerts, err
}
