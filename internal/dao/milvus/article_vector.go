package milvus

import (
	"context"
	"fmt"
	"time"

	"techmind/internal/monitor"
	"techmind/internal/pkg/settings"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// ArticleChunkVector 表示一个 chunk 的向量数据
type ArticleChunkVector struct {
	ID         int64
	ArticleID  int64
	ChunkIndex int32
	Embedding  []float32
}

// InsertArticleChunks 批量写入文章 chunk 向量
func InsertArticleChunks(ctx context.Context, collectionName string, chunks []ArticleChunkVector) error {
	if len(chunks) == 0 {
		return nil
	}

	ids := make([]int64, len(chunks))
	articleIDs := make([]int64, len(chunks))
	chunkIndexes := make([]int32, len(chunks))
	embeddings := make([][]float32, len(chunks))

	for i, c := range chunks {
		ids[i] = c.ID
		articleIDs[i] = c.ArticleID
		chunkIndexes[i] = c.ChunkIndex
		embeddings[i] = c.Embedding
	}

	columns := []entity.Column{
		entity.NewColumnInt64("id", ids),
		entity.NewColumnInt64("article_id", articleIDs),
		entity.NewColumnInt32("chunk_index", chunkIndexes),
		entity.NewColumnFloatVector("embedding", len(embeddings[0]), embeddings),
	}

	_, err := Client.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("milvus: insert chunks: %w", err)
	}

	return Client.Flush(ctx, collectionName, false)
}

// DeleteArticleChunks 按 article_id 删除该文章所有 chunk 向量
func DeleteArticleChunks(ctx context.Context, collectionName string, articleID int64) error {
	expr := fmt.Sprintf("article_id == %d", articleID)
	return Client.Delete(ctx, collectionName, "", expr)
}

// SearchSimilar 语义搜索，返回最相似的 topK 个 article_id（可能重复，需去重）
func SearchSimilar(ctx context.Context, cfg *settings.MilvusSetting, queryVec []float32, topK int) ([]int64, error) {
	sp, err := entity.NewIndexHNSWSearchParam(64)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	results, err := Client.Search(
		ctx,
		cfg.CollectionName,
		nil,
		"",
		[]string{"article_id"},
		[]entity.Vector{entity.FloatVector(queryVec)},
		"embedding",
		entity.L2,
		topK,
		sp,
	)
	monitor.ObserveMilvusSearch(time.Since(start), err)
	if err != nil {
		return nil, fmt.Errorf("milvus: search: %w", err)
	}

	seen := make(map[int64]struct{})
	var articleIDs []int64
	for _, result := range results {
		col := result.Fields.GetColumn("article_id")
		if col == nil {
			continue
		}
		colInt64, ok := col.(*entity.ColumnInt64)
		if !ok {
			continue
		}
		for i := 0; i < col.Len(); i++ {
			val, err := colInt64.ValueByIdx(i)
			if err != nil {
				continue
			}
			if _, dup := seen[val]; !dup {
				seen[val] = struct{}{}
				articleIDs = append(articleIDs, val)
			}
		}
	}
	return articleIDs, nil
}
