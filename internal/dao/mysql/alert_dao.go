package mysql

import (
	"errors"
	"time"

	"techmind/internal/model"

	"gorm.io/gorm"
)

// UpsertAlertEvent 写入或去重更新告警事件
// fingerprint 唯一：已存在则更新 repeat_count/last_seen_at/status；否则插入
func UpsertAlertEvent(event *model.AlertEvent) error {
	var exists model.AlertEvent
	err := DB.Where("fingerprint = ?", event.Fingerprint).First(&exists).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 首次收到
		return DB.Create(event).Error
	}
	if err != nil {
		return err
	}

	// 已存在：更新重复计数和最近时间
	updates := map[string]interface{}{
		"repeat_count": gorm.Expr("repeat_count + 1"),
		"last_seen_at": event.LastSeenAt,
		"status":       event.Status,
	}
	if event.Status == model.AlertStatusResolved {
		now := time.Now()
		updates["resolved_at"] = now
	}
	if err := DB.Model(&exists).Updates(updates).Error; err != nil {
		return err
	}
	// 把生成的 id 回填，方便调用方使用
	event.ID = exists.ID
	return nil
}

// GetAlertEventByID 按 ID 查询
func GetAlertEventByID(id int64) (*model.AlertEvent, error) {
	var e model.AlertEvent
	err := DB.First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &e, err
}

// ListAlertEvents 分页查询告警列表（按最近时间倒序）
func ListAlertEvents(status string, page, pageSize int) ([]*model.AlertEvent, int, error) {
	offset := (page - 1) * pageSize

	q := DB.Model(&model.AlertEvent{})
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.AlertEvent
	err := q.Order("last_seen_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// AcknowledgeAlert 将告警状态改为 acknowledged
func AcknowledgeAlert(id int64) error {
	return DB.Model(&model.AlertEvent{}).Where("id = ?", id).
		Update("status", model.AlertStatusAcknowledged).Error
}

// CreateAlertEnrichment 写入告警增强上下文
func CreateAlertEnrichment(e *model.AlertEnrichment) error {
	return DB.Create(e).Error
}

// GetAlertEnrichmentByAlertID 查询告警增强内容
func GetAlertEnrichmentByAlertID(alertID int64) (*model.AlertEnrichment, error) {
	var e model.AlertEnrichment
	err := DB.Where("alert_id = ?", alertID).
		Order("created_at DESC").First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &e, err
}
