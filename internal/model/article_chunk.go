package model

// ArticleChunk 对应 article_chunk 表
type ArticleChunk struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID  int64  `gorm:"not null;index"           json:"article_id,string"`
	ChunkIndex int    `gorm:"default:0"                json:"chunk_index"`
	Content    string `gorm:"type:text;not null"       json:"content"`
	MilvusID   string `gorm:"default:''"               json:"milvus_id"`
}

func (ArticleChunk) TableName() string { return "article_chunk" }
