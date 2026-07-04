package milvus

import (
	"context"
	"fmt"

	"techmind/internal/pkg/settings"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// RunbookChunkVector Runbook chunk 向量数据
type RunbookChunkVector struct {
	ID         int64
	RunbookID  int64
	ChunkIndex int32
	Embedding  []float32
}

// InsertRunbookChunks 写入 Runbook chunk 向量
// 使用负数 article_id（-runbookID）与文章区分
func InsertRunbookChunks(ctx context.Context, collectionName string, chunks []RunbookChunkVector) error {
	if len(chunks) == 0 {
		return nil
	}

	ids := make([]int64, len(chunks))
	articleIDs := make([]int64, len(chunks)) // 负数存 Runbook ID
	chunkIndexes := make([]int32, len(chunks))
	embeddings := make([][]float32, len(chunks))

	for i, c := range chunks {
		ids[i] = c.ID
		articleIDs[i] = -c.RunbookID // 负数区分文章和 Runbook
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
		return fmt.Errorf("milvus: insert runbook chunks: %w", err)
	}
	return Client.Flush(ctx, collectionName, false)
}

// DeleteRunbookChunks 按 runbook_id（负数存储形式）删除该 Runbook 的所有向量
func DeleteRunbookChunks(ctx context.Context, collectionName string, runbookID int64) error {
	expr := fmt.Sprintf("article_id == %d", -runbookID)
	return Client.Delete(ctx, collectionName, "", expr)
}

// SearchRunbookSimilar 只在 Runbook 向量中检索（article_id < 0 的记录）
func SearchRunbookSimilar(ctx context.Context, cfg *settings.MilvusSetting, queryVec []float32, topK int) ([]int64, error) {
	sp, err := entity.NewIndexHNSWSearchParam(64)
	if err != nil {
		return nil, err
	}

	results, err := Client.Search(
		ctx,
		cfg.CollectionName,
		nil,
		"article_id < 0", // 只检索 Runbook
		[]string{"article_id"},
		[]entity.Vector{entity.FloatVector(queryVec)},
		"embedding",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("milvus: runbook search: %w", err)
	}

	seen := make(map[int64]struct{})
	var runbookIDs []int64
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
			rbID := -val // 负数取反得到 Runbook ID
			if _, dup := seen[rbID]; !dup {
				seen[rbID] = struct{}{}
				runbookIDs = append(runbookIDs, rbID)
			}
		}
	}
	return runbookIDs, nil
}
