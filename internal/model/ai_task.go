package model

import "time"

// AITask 对应 ai_task 表，记录异步任务状态
type AITask struct {
	ID         int64     `gorm:"primaryKey"`
	TaskType   string    `gorm:"not null;index"`
	RefID      int64     `gorm:"not null;index"`
	Status     string    `gorm:"default:'pending'"` // pending/running/done/failed/dead
	RetryCount int       `gorm:"default:0"`
	ErrorMsg   string    `gorm:"default:''"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (AITask) TableName() string { return "ai_task" }

const (
	AITaskStatusPending = "pending"
	AITaskStatusRunning = "running"
	AITaskStatusDone    = "done"
	AITaskStatusFailed  = "failed"
	AITaskStatusDead    = "dead"
)
