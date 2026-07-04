package mysql

import (
	"techmind/internal/model"
)

// CreateAICallRecord 写入 AI 调用记录
func CreateAICallRecord(r *model.AICallRecord) error {
	return DB.Create(r).Error
}

// ListAICallRecords 分页查询 AI 调用记录（按时间倒序）
func ListAICallRecords(skill string, page, pageSize int) ([]*model.AICallRecord, int, error) {
	offset := (page - 1) * pageSize
	q := DB.Model(&model.AICallRecord{})
	if skill != "" {
		q = q.Where("skill = ?", skill)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.AICallRecord
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, int(total), err
}
