package model

import "time"

// Tag 对应 tag 表
type Tag struct {
	ID        int64     `gorm:"primaryKey"   json:"id,string"`
	Name      string    `gorm:"uniqueIndex"  json:"name"`
	HotScore  float64   `gorm:"default:0"    json:"hot_score"`
	CreatedAt time.Time `                    json:"created_at"`
	UpdatedAt time.Time `                    json:"updated_at"`
}

func (Tag) TableName() string { return "tag" }

// ArticleTag 对应 article_tag 表
type ArticleTag struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	ArticleID int64     `gorm:"index;not null"`
	TagID     int64     `gorm:"index;not null"`
	Source    string    `gorm:"default:'manual'"` // manual / ai
	CreatedAt time.Time
}

func (ArticleTag) TableName() string { return "article_tag" }
