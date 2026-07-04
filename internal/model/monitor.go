package model

import "time"

// MonitorSlowRequest 对应 monitor_slow_request 表
type MonitorSlowRequest struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID  string    `json:"request_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	DurationMs int       `json:"duration_ms"`
	UserID     int64     `json:"user_id,string"`
	CreatedAt  time.Time `json:"created_at"`
}

func (MonitorSlowRequest) TableName() string { return "monitor_slow_request" }

// MonitorErrorEvent 对应 monitor_error_event 表
type MonitorErrorEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Source    string    `json:"source"` // panic/mysql/redis/milvus/ai/http/worker
	Path      string    `json:"path"`
	RequestID string    `json:"request_id"`
	Message   string    `json:"message"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MonitorErrorEvent) TableName() string { return "monitor_error_event" }
