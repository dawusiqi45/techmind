package model

import "time"

// AICallRecord 对应 ai_call_record 表，记录每次 LLM/Embedding 调用详情
type AICallRecord struct {
	ID           int64     `gorm:"primaryKey"    json:"id,string"`
	Skill        string    `gorm:"not null;index" json:"skill"` // llm / embedding / embedding_batch / ops_diagnose 等
	Model        string    `gorm:"default:''"    json:"model"`
	InputTokens  int       `gorm:"default:0"     json:"input_tokens"`
	OutputTokens int       `gorm:"default:0"     json:"output_tokens"`
	DurationMs   int       `gorm:"default:0"     json:"duration_ms"`
	Status       string    `gorm:"default:'ok'"  json:"status"` // ok / failed
	ErrorMsg     string    `gorm:"default:''"    json:"error_msg"`
	RefID        int64     `gorm:"default:0"     json:"ref_id,string"` // 关联文章/报告 ID
	CreatedAt    time.Time `                     json:"created_at"`
}

func (AICallRecord) TableName() string { return "ai_call_record" }
