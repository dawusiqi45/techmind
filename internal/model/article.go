package model

import "time"

// Article 对应 article 表
type Article struct {
	ID            int64     `gorm:"primaryKey"     json:"id,string"`
	AuthorID      int64     `gorm:"index;not null" json:"author_id,string"`
	Title         string    `gorm:"not null"       json:"title"`
	Content       string    `gorm:"not null"       json:"content"`
	Summary       string    `gorm:"default:''"    json:"summary"`
	Cover         string    `gorm:"default:''"    json:"cover"`
	Status        int8      `gorm:"default:1;index" json:"status"`
	IndexStatus   int8      `gorm:"column:index_status;default:0" json:"index_status"`
	ViewCount     int       `gorm:"default:0"     json:"view_count"`
	LikeCount     int       `gorm:"default:0"     json:"like_count"`
	FavoriteCount int       `gorm:"default:0"     json:"favorite_count"`
	CommentCount  int       `gorm:"default:0"     json:"comment_count"`
	CreatedAt     time.Time `                     json:"created_at"`
	UpdatedAt     time.Time `                     json:"updated_at"`
}

func (Article) TableName() string { return "article" }

// ArticleDetail 文章详情，附带作者信息和标签
type ArticleDetail struct {
	Article
	AuthorName string   `gorm:"column:author_name" json:"author_name"`
	Tags       []string `gorm:"-"                  json:"tags"`
}

// ArticleListItem 文章列表摘要，不含正文，用于 JOIN 查询 Scan
type ArticleListItem struct {
	ID            int64     `gorm:"column:id"             json:"id,string"`
	AuthorID      int64     `gorm:"column:author_id"      json:"author_id,string"`
	AuthorName    string    `gorm:"column:author_name"    json:"author_name"`
	Title         string    `gorm:"column:title"          json:"title"`
	Summary       string    `gorm:"column:summary"        json:"summary"`
	Cover         string    `gorm:"column:cover"          json:"cover"`
	ViewCount     int       `gorm:"column:view_count"     json:"view_count"`
	LikeCount     int       `gorm:"column:like_count"     json:"like_count"`
	FavoriteCount int       `gorm:"column:favorite_count" json:"favorite_count"`
	CommentCount  int       `gorm:"column:comment_count"  json:"comment_count"`
	CreatedAt     time.Time `gorm:"column:created_at"     json:"created_at"`
}
