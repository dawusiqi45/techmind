package milvus

import (
	"context"
	"fmt"

	"techmind/internal/pkg/settings"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// Client 是全局 Milvus 客户端
var Client client.Client

// Init 连接 Milvus 并确保 Collection 存在
func Init(cfg *settings.MilvusSetting) error {
	c, err := client.NewClient(context.Background(), client.Config{
		Address: cfg.Addr,
	})
	if err != nil {
		return fmt.Errorf("milvus: connect failed: %w", err)
	}
	Client = c

	if err := ensureCollection(context.Background(), cfg); err != nil {
		return fmt.Errorf("milvus: ensure collection: %w", err)
	}
	return nil
}

// Close 关闭 Milvus 连接
func Close() {
	if Client != nil {
		_ = Client.Close()
	}
}

// ensureCollection 如果 collection 不存在则创建
func ensureCollection(ctx context.Context, cfg *settings.MilvusSetting) error {
	exists, err := Client.HasCollection(ctx, cfg.CollectionName)
	if err != nil {
		return err
	}
	if exists {
		// 确保已加载到内存（Milvus 查询前需要 Load）
		_ = Client.LoadCollection(ctx, cfg.CollectionName, false)
		return nil
	}

	schema := &entity.Schema{
		CollectionName: cfg.CollectionName,
		Description:    "TechMind article chunks for semantic search",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     false,
			},
			{
				Name:     "article_id",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:     "chunk_index",
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", cfg.Dim),
				},
			},
		},
	}

	if err := Client.CreateCollection(ctx, schema, 1); err != nil {
		return err
	}

	// 创建向量索引（HNSW）
	idx, err := entity.NewIndexHNSW(entity.L2, 8, 64)
	if err != nil {
		return err
	}
	if err := Client.CreateIndex(ctx, cfg.CollectionName, "embedding", idx, false); err != nil {
		return err
	}

	return Client.LoadCollection(ctx, cfg.CollectionName, false)
}
