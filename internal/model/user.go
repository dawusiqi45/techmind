package model

import "time"

// User 对应 user 表
type User struct {
	ID        int64     `gorm:"primaryKey"                json:"-"`
	Username  string    `gorm:"uniqueIndex;not null"      json:"username"`
	Password  string    `gorm:"not null"                  json:"-"`
	Email     string    `gorm:"uniqueIndex"               json:"email"`
	Avatar    string    `                                 json:"avatar"`
	Role      int8      `gorm:"default:0"                 json:"role"`
	Status    int8      `gorm:"default:1"                 json:"status"`
	CreatedAt time.Time `                                 json:"created_at"`
	UpdatedAt time.Time `                                 json:"updated_at"`
}

func (User) TableName() string { return "user" }

// UserProfile 用于对外返回，不含密码
type UserProfile struct {
	ID        int64     `json:"id,string"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Role      int8      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UserLike struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index"          json:"user_id"`
	ArticleID int64     `gorm:"not null"                 json:"article_id"`
	CreatedAt time.Time `                                json:"created_at"`
}

func (UserLike) TableName() string { return "user_like" }
