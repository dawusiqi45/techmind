package model

import "time"

// Runbook 对应 runbook 表
type Runbook struct {
	ID          int64     `gorm:"primaryKey"       json:"id,string"`
	Title       string    `gorm:"not null"         json:"title"`
	Content     string    `gorm:"type:mediumtext"  json:"content"`
	AlertName   string    `gorm:"default:''"`      // 关联的告警名称，可为空
	Service     string    `gorm:"default:''"`
	IndexStatus int8      `gorm:"default:0"` // 0=未索引 1=已索引
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Runbook) TableName() string { return "runbook" }
