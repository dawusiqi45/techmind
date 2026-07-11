package mysql

import (
	"errors"
	"fmt"

	"techmind/internal/model"
	"techmind/internal/pkg/snowflake"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// EnsureOpenIncidentForAlert 返回告警当前关联的开放故障事件；没有时会创建一个。
// 对 alert_event 行加锁可避免同一告警的重试任务并发创建多个 Incident。
func EnsureOpenIncidentForAlert(alertID int64) (*model.Incident, error) {
	var result *model.Incident
	err := DB.Transaction(func(tx *gorm.DB) error {
		var alert model.AlertEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&alert, alertID).Error; err != nil {
			return err
		}

		var incident model.Incident
		query := tx.Table("incident AS i").
			Select("i.*").
			Joins("JOIN incident_alert AS ia ON ia.incident_id = i.id").
			Where("ia.alert_id = ? AND i.status = ?", alertID, "open").
			Order("i.created_at DESC").
			Limit(1).
			Find(&incident)
		if query.Error != nil {
			return query.Error
		}
		if query.RowsAffected > 0 {
			result = &incident
			return nil
		}

		title := alert.AlertName
		if alert.Service != "" {
			title = fmt.Sprintf("%s · %s", alert.AlertName, alert.Service)
		}
		incident = model.Incident{
			ID:       snowflake.GenID(),
			Title:    title,
			Status:   "open",
			Severity: alert.Severity,
		}
		if err := tx.Create(&incident).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.IncidentAlert{IncidentID: incident.ID, AlertID: alertID}).Error; err != nil {
			return err
		}
		result = &incident
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveIncident 将故障事件标记为已解决；它不会修改任何告警状态。
func ResolveIncident(id int64) error {
	return DB.Model(&model.Incident{}).Where("id = ?", id).Update("status", "resolved").Error
}
