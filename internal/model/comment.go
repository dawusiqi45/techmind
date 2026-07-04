package model

import "time"

// Comment 对应 comment 表
type Comment struct {
	ID        int64     `gorm:"primaryKey"     json:"id,string"`
	ArticleID int64     `gorm:"index;not null" json:"article_id,string"`
	AuthorID  int64     `gorm:"not null"       json:"author_id,string"`
	ParentID  int64     `gorm:"default:0"      json:"parent_id,string"`
	Content   string    `gorm:"not null"       json:"content"`
	Status    int8      `gorm:"default:1"      json:"status"`
	CreatedAt time.Time `                      json:"created_at"`
	UpdatedAt time.Time `                      json:"updated_at"`
}

func (Comment) TableName() string { return "comment" }

// CommentDetail 评论详情，附带作者信息（JOIN 查询 Scan 使用）
type CommentDetail struct {
	Comment
	AuthorName   string           `gorm:"column:author_name"   json:"author_name"`
	AuthorAvatar string           `gorm:"column:author_avatar" json:"author_avatar"`
	Replies      []*CommentDetail `gorm:"-"                    json:"replies,omitempty"`
}
