package mysql

import (
	"context"
	"time"

	"techmind/internal/model"

	"gorm.io/gorm"
)

// CreateSlowRequest 写入慢请求记录
func CreateSlowRequest(req *model.MonitorSlowRequest) error {
	return DB.Create(req).Error
}

// CreateErrorEvent 写入错误事件；相同 source/path/message 进行聚合计数
func CreateErrorEvent(event *model.MonitorErrorEvent) error {
	var exists model.MonitorErrorEvent
	err := DB.Where("source = ? AND path = ? AND message = ?", event.Source, event.Path, event.Message).
		First(&exists).Error
	if err == nil {
		return DB.Model(&exists).Updates(map[string]interface{}{
			"count":      gorm.Expr("count + 1"),
			"request_id": event.RequestID,
		}).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	event.Count = 1
	return DB.Create(event).Error
}

// ListSlowRequests 分页查询慢请求（按时间倒序）
func ListSlowRequests(page, pageSize int) ([]*model.MonitorSlowRequest, int, error) {
	offset := (page - 1) * pageSize
	var total int64
	if err := DB.Model(&model.MonitorSlowRequest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.MonitorSlowRequest
	err := DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// ListSlowRequestsInWindow 查询诊断时间窗内的慢请求，避免混入无关历史样本。
func ListSlowRequestsInWindow(ctx context.Context, start, end time.Time, limit int) ([]*model.MonitorSlowRequest, int, error) {
	q := DB.WithContext(ctx).Model(&model.MonitorSlowRequest{}).
		Where("created_at BETWEEN ? AND ?", start, end)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.MonitorSlowRequest
	err := q.Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, int(total), err
}

// ListErrorEvents 分页查询错误事件（按更新时间倒序）
func ListErrorEvents(source string, page, pageSize int) ([]*model.MonitorErrorEvent, int, error) {
	offset := (page - 1) * pageSize
	q := DB.Model(&model.MonitorErrorEvent{})
	if source != "" {
		q = q.Where("source = ?", source)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.MonitorErrorEvent
	err := q.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}

// ListErrorEventsInWindow 查询时间窗内有更新的聚合错误事件。
func ListErrorEventsInWindow(ctx context.Context, source string, start, end time.Time, limit int) ([]*model.MonitorErrorEvent, int, error) {
	q := DB.WithContext(ctx).Model(&model.MonitorErrorEvent{}).
		Where("updated_at BETWEEN ? AND ?", start, end)
	if source != "" {
		q = q.Where("source = ?", source)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.MonitorErrorEvent
	err := q.Order("updated_at DESC").Limit(limit).Find(&list).Error
	return list, int(total), err
}
