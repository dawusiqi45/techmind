package ai

import (
	"context"
	"fmt"
	"time"

	"techmind/internal/monitor"
	"techmind/internal/pkg/settings"

	arkEmbedding "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
)

// embeddingClient 全局 Embedding 客户端（懒初始化）
var embeddingClient embedding.Embedder

// InitEmbedding 初始化 Ark Embedding 客户端
func InitEmbedding(cfg *settings.AISetting) error {
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	c, err := arkEmbedding.NewEmbedder(context.Background(), &arkEmbedding.EmbeddingConfig{
		APIKey:  cfg.EmbeddingAPIKey,
		Model:   cfg.EmbeddingModel,
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("ai: init embedding: %w", err)
	}
	embeddingClient = c
	return nil
}

// EmbedText 对单段文本生成 embedding 向量
func EmbedText(ctx context.Context, text string) ([]float32, error) {
	if embeddingClient == nil {
		return nil, fmt.Errorf("ai: embedding client not initialized")
	}
	start := time.Now()
	vecs, err := embeddingClient.EmbedStrings(ctx, []string{text})
	monitor.ObserveAICall("embedding", time.Since(start), err)
	if err != nil {
		return nil, fmt.Errorf("ai: embed text: %w", err)
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("ai: embed returned empty result")
	}
	result := make([]float32, len(vecs[0]))
	for i, v := range vecs[0] {
		result[i] = float32(v)
	}
	return result, nil
}

// EmbedBatch 批量生成 embedding，每段文本对应一个 []float32
func EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if embeddingClient == nil {
		return nil, fmt.Errorf("ai: embedding client not initialized")
	}
	start := time.Now()
	vecs, err := embeddingClient.EmbedStrings(ctx, texts)
	monitor.ObserveAICall("embedding_batch", time.Since(start), err)
	if err != nil {
		return nil, fmt.Errorf("ai: embed batch: %w", err)
	}
	results := make([][]float32, len(vecs))
	for i, vec := range vecs {
		f32 := make([]float32, len(vec))
		for j, v := range vec {
			f32[j] = float32(v)
		}
		results[i] = f32
	}
	return results, nil
}
