package worker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"techmind/internal/agent"
	aiEmbed "techmind/internal/ai/embedding"
	aiSkill "techmind/internal/ai/prompt"
	milvusDAO "techmind/internal/dao/milvus"
	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/model"
	"techmind/internal/pkg/settings"
	"techmind/internal/pkg/snowflake"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RegisterAIHandlers 将所有 AI 任务处理器注册到 Worker
func RegisterAIHandlers(w *AIWorker) {
	w.Register(redisDAO.TaskArticleSummary, handleArticleSummary)
	w.Register(redisDAO.TaskArticleTag, handleArticleTag)
	w.Register(redisDAO.TaskArticleIndex, handleArticleIndex)
	w.Register(redisDAO.TaskArticleReindex, handleArticleReindex)
	w.Register(redisDAO.TaskArticleDeleteIndex, handleArticleDeleteIndex)
	w.Register(redisDAO.TaskRunbookIndex, handleRunbookIndex)
}

func handleRunbookIndex(ctx context.Context, msg goredis.XMessage) error {
	runbookID, err := extractRefID(msg)
	if err != nil {
		return err
	}
	runbook, err := mysqlDAO.GetRunbookByID(runbookID)
	if err != nil || runbook == nil {
		return fmt.Errorf("runbook index: runbook %d not found", runbookID)
	}
	if err := agent.IndexRunbook(ctx, runbook.ID, runbook.Title, runbook.Content); err != nil {
		_ = mysqlDAO.UpdateRunbookIndexStatus(runbookID, -1)
		return err
	}
	if err := mysqlDAO.UpdateRunbookIndexStatus(runbookID, 1); err != nil {
		return fmt.Errorf("runbook index: update status: %w", err)
	}
	return nil
}

// handleArticleSummary 调用 LLM 生成文章摘要并回写 DB
func handleArticleSummary(ctx context.Context, msg goredis.XMessage) error {
	articleID, err := extractRefID(msg)
	if err != nil {
		return err
	}

	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil || a == nil {
		return fmt.Errorf("summary: article %d not found", articleID)
	}

	summary, err := aiSkill.ArticleSummarySkill(ctx, a.Title, a.Content)
	if err != nil {
		return fmt.Errorf("summary: skill failed: %w", err)
	}

	if err := mysqlDAO.UpdateArticleSummary(articleID, summary); err != nil {
		return fmt.Errorf("summary: update db: %w", err)
	}
	zap.L().Info("task: article.summary done", zap.Int64("article_id", articleID))
	return nil
}

// handleArticleTag 调用 LLM 生成 AI 标签并写入 article_tag
func handleArticleTag(ctx context.Context, msg goredis.XMessage) error {
	articleID, err := extractRefID(msg)
	if err != nil {
		return err
	}

	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil || a == nil {
		return fmt.Errorf("tag: article %d not found", articleID)
	}

	tagNames, err := aiSkill.ArticleTagSkill(ctx, a.Title, a.Content)
	if err != nil {
		return fmt.Errorf("tag: skill failed: %w", err)
	}

	var tagIDs []int64
	for _, name := range tagNames {
		tagID, err := mysqlDAO.GetOrCreateTag(name, snowflake.GenID())
		if err != nil {
			zap.L().Warn("tag: get or create failed", zap.String("name", name), zap.Error(err))
			continue
		}
		tagIDs = append(tagIDs, tagID)
	}
	if err := mysqlDAO.UpsertArticleTags(articleID, tagIDs, "ai"); err != nil {
		return fmt.Errorf("tag: upsert: %w", err)
	}
	zap.L().Info("task: article.tag done",
		zap.Int64("article_id", articleID),
		zap.String("tags", strings.Join(tagNames, ",")))
	return nil
}

// handleArticleIndex 首次索引：切分 chunk → embedding → 写入 Milvus
func handleArticleIndex(ctx context.Context, msg goredis.XMessage) error {
	articleID, err := extractRefID(msg)
	if err != nil {
		return err
	}
	return indexArticle(ctx, articleID, false)
}

// handleArticleReindex 重新索引：先删再建
func handleArticleReindex(ctx context.Context, msg goredis.XMessage) error {
	articleID, err := extractRefID(msg)
	if err != nil {
		return err
	}
	return indexArticle(ctx, articleID, true)
}

// handleArticleDeleteIndex 删除文章的所有向量
func handleArticleDeleteIndex(ctx context.Context, msg goredis.XMessage) error {
	articleID, err := extractRefID(msg)
	if err != nil {
		return err
	}
	col := settings.Conf.Milvus.CollectionName
	if err := milvusDAO.DeleteArticleChunks(ctx, col, articleID); err != nil {
		return fmt.Errorf("delete_index: %w", err)
	}
	zap.L().Info("task: article.delete_index done", zap.Int64("article_id", articleID))
	return nil
}

// indexArticle 执行向量索引（reindex=true 先删旧向量）
func indexArticle(ctx context.Context, articleID int64, reindex bool) error {
	a, err := mysqlDAO.GetArticleByID(articleID)
	if err != nil || a == nil {
		return fmt.Errorf("index: article %d not found", articleID)
	}

	col := settings.Conf.Milvus.CollectionName

	if reindex {
		_ = milvusDAO.DeleteArticleChunks(ctx, col, articleID)
		// 同步清理 MySQL chunk 记录
		_ = mysqlDAO.DeleteArticleChunks(articleID)
	}

	// 切分 Markdown（按段落，每段最多 500 字）
	chunks := splitMarkdown(a.Content, 500)
	if len(chunks) == 0 {
		return nil
	}

	// 批量生成 embedding
	embeddings, err := aiEmbed.EmbedBatch(ctx, chunks)
	if err != nil {
		return fmt.Errorf("index: embed batch: %w", err)
	}

	// 构造 Milvus 写入数据，同时构建 MySQL chunk 记录
	vectors := make([]milvusDAO.ArticleChunkVector, len(chunks))
	sqlChunks := make([]*model.ArticleChunk, len(chunks))
	for i, emb := range embeddings {
		chunkID := snowflake.GenID()
		vectors[i] = milvusDAO.ArticleChunkVector{
			ID:         chunkID,
			ArticleID:  articleID,
			ChunkIndex: int32(i),
			Embedding:  emb,
		}
		sqlChunks[i] = &model.ArticleChunk{
			ArticleID:  articleID,
			ChunkIndex: i,
			Content:    chunks[i],
			MilvusID:   strconv.FormatInt(chunkID, 10),
		}
	}

	if err := milvusDAO.InsertArticleChunks(ctx, col, vectors); err != nil {
		return fmt.Errorf("index: insert: %w", err)
	}

	// 同步写 MySQL article_chunk 表
	_ = mysqlDAO.BatchCreateArticleChunks(sqlChunks)

	zap.L().Info("task: article.index done",
		zap.Int64("article_id", articleID),
		zap.Int("chunks", len(chunks)))
	return nil
}

// splitMarkdown 按段落切分 Markdown 文本，每段不超过 maxLen 字符
func splitMarkdown(content string, maxLen int) []string {
	paragraphs := strings.Split(content, "\n\n")
	var chunks []string
	var buf strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if buf.Len()+len([]rune(p)) > maxLen && buf.Len() > 0 {
			chunks = append(chunks, buf.String())
			buf.Reset()
		}
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(p)
	}
	if buf.Len() > 0 {
		chunks = append(chunks, buf.String())
	}
	return chunks
}

// extractRefID 从 Stream 消息中解析 ref_id（文章 ID 等）
func extractRefID(msg goredis.XMessage) (int64, error) {
	v, ok := msg.Values["ref_id"]
	if !ok {
		return 0, fmt.Errorf("missing ref_id in message %s", msg.ID)
	}
	id, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ref_id %v: %w", v, err)
	}
	return id, nil
}
