package model

import "time"

// Favorite 对应 favorite 表
type Favorite struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"uniqueIndex:uk_user_article;not null"`
	ArticleID int64     `gorm:"uniqueIndex:uk_user_article;not null"`
	CreatedAt time.Time
}

func (Favorite) TableName() string { return "favorite" }
