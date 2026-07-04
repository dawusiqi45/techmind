package model

import "time"

// Incident 对应 incident 表，关联多条 AlertEvent 形成故障事件
type Incident struct {
	ID        int64     `gorm:"primaryKey"            json:"id,string"`
	Title     string    `gorm:"not null"              json:"title"`
	Status    string    `gorm:"default:'open';index"  json:"status"` // open/resolved
	Severity  string    `gorm:"default:'warning'"     json:"severity"`
	CreatedAt time.Time `                             json:"created_at"`
	UpdatedAt time.Time `                             json:"updated_at"`
}

func (Incident) TableName() string { return "incident" }

// IncidentAlert 对应 incident_alert 关联表
type IncidentAlert struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	IncidentID int64     `gorm:"uniqueIndex:uk_incident_alert;not null" json:"incident_id,string"`
	AlertID    int64     `gorm:"uniqueIndex:uk_incident_alert;not null" json:"alert_id,string"`
	CreatedAt  time.Time
}

func (IncidentAlert) TableName() string { return "incident_alert" }
