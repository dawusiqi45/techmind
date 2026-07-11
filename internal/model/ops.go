package model

import "time"

// OpsReport 对应 ops_report 表，SRE Agent 生成的诊断报告
type OpsReport struct {
	ID             int64     `gorm:"primaryKey"                   json:"id,string"`
	AlertID        int64     `gorm:"default:0;index"              json:"alert_id,string"`
	IncidentID     int64     `gorm:"default:0;index"              json:"incident_id,string"`
	TriggerType    string    `gorm:"default:'manual'"             json:"trigger_type"` // manual/alert
	Summary        string    `gorm:"type:text"                    json:"summary"`
	Evidence       JSONSlice `gorm:"serializer:json"              json:"evidence"`
	RootCause      string    `gorm:"type:text"                    json:"root_cause"`
	Impact         string    `gorm:"type:text"                    json:"impact"`
	Suggestions    JSONSlice `gorm:"serializer:json"              json:"suggestions"`
	RelatedChanges JSONSlice `gorm:"serializer:json"              json:"related_changes"`
	ToolCalls      JSONSlice `gorm:"serializer:json"              json:"tool_calls"`
	Status         string    `gorm:"default:'done'"               json:"status"` // running/done/failed
	CreatedAt      time.Time `                                    json:"created_at"`
}

func (OpsReport) TableName() string { return "ops_report" }

// OpsToolCall 对应 ops_tool_call 表，保存诊断中每次真实的只读工具调用。
type OpsToolCall struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"      json:"id,string"`
	ReportID   int64     `gorm:"not null;index"                json:"report_id,string"`
	ToolName   string    `gorm:"not null"                      json:"tool_name"`
	Input      JSONMap   `gorm:"serializer:json"               json:"input"`
	Output     JSONMap   `gorm:"serializer:json"               json:"output"`
	DurationMs int       `gorm:"default:0"                     json:"duration_ms"`
	CreatedAt  time.Time `                                      json:"created_at"`
}

func (OpsToolCall) TableName() string { return "ops_tool_call" }

// JSONSlice 是 JSON 数组列的 Go 类型
type JSONSlice []interface{}

// DeploymentChange 对应 deployment_change 表
type DeploymentChange struct {
	ID        int64     `gorm:"primaryKey"            json:"id,string"`
	Service   string    `gorm:"not null;index"        json:"service"`
	Namespace string    `gorm:"default:'default'"     json:"namespace"`
	Image     string    `gorm:"default:''"            json:"image"`
	OldImage  string    `gorm:"default:''"            json:"old_image"`
	Replicas  int       `gorm:"default:0"             json:"replicas"`
	ChangedBy string    `gorm:"default:''"            json:"changed_by"`
	Source    string    `gorm:"default:'manual'"      json:"source"` // helm/kubectl/argocd/manual
	ChangedAt time.Time `gorm:"index"                 json:"changed_at"`
	CreatedAt time.Time `                             json:"created_at"`
}

func (DeploymentChange) TableName() string { return "deployment_change" }
