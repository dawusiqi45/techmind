package model

import "time"

// AlertEvent 对应 alert_event 表
type AlertEvent struct {
	ID          int64      `gorm:"primaryKey"                    json:"id,string"`
	Fingerprint string     `gorm:"uniqueIndex;not null"          json:"fingerprint"`
	AlertName   string     `gorm:"not null;index"                json:"alert_name"`
	Service     string     `gorm:"default:''"                    json:"service"`
	Endpoint    string     `gorm:"default:''"                    json:"endpoint"`
	Severity    string     `gorm:"default:'warning'"             json:"severity"`
	Status      string     `gorm:"default:'firing';index"        json:"status"` // firing/acknowledged/resolved
	Labels      JSONMap    `gorm:"serializer:json"               json:"labels"`
	Annotations JSONMap    `gorm:"serializer:json"               json:"annotations"`
	RepeatCount int        `gorm:"default:1"                     json:"repeat_count"`
	FirstSeenAt time.Time  `                                     json:"first_seen_at"`
	LastSeenAt  time.Time  `                                     json:"last_seen_at"`
	ResolvedAt  *time.Time `                                     json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `                                     json:"created_at"`
	UpdatedAt   time.Time  `                                     json:"updated_at"`
}

func (AlertEvent) TableName() string { return "alert_event" }

// JSONMap 是 JSON 列的 Go 类型
type JSONMap map[string]interface{}

// AlertEnrichment 对应 alert_enrichment 表
type AlertEnrichment struct {
	ID        int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	AlertID   int64   `gorm:"index;not null"           json:"alert_id,string"`
	Context   JSONMap `gorm:"serializer:json"          json:"context"`
	CreatedAt time.Time
}

func (AlertEnrichment) TableName() string { return "alert_enrichment" }

// AlertStatusFiring 告警状态常量
const (
	AlertStatusFiring       = "firing"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved     = "resolved"
)
