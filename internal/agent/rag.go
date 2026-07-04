package agent

import (
	"context"
	"fmt"
	"strings"

	milvusDAO "techmind/internal/dao/milvus"
	mysqlDAO "techmind/internal/dao/mysql"
	aiEmbed "techmind/internal/ai/embedding"
	"techmind/internal/pkg/settings"
	"techmind/internal/pkg/snowflake"
)

// RAGResult RAG 检索结果
type RAGResult struct {
	RunbookSummaries []string // 相关 Runbook 摘要
	ReportSummaries  []string // 相关历史报告摘要
}

// IncidentRAGSkill 根据告警名称和服务检索相关 Runbook 和历史诊断报告
// 1. 用关键词检索 MySQL 中关联同名告警的 Runbook（精确召回）
// 2. 用语义向量在 Milvus 检索最相似的 Runbook chunk（语义召回）
// 3. 检索历史 ops_report 中近似场景
func IncidentRAGSkill(ctx context.Context, alertName, service, query string) (*RAGResult, error) {
	result := &RAGResult{}

	// ── 1. MySQL 精确召回 Runbook ──────────────────────────
	exactBooks, err := mysqlDAO.SearchRunbooksByAlertName(alertName, service, 3)
	if err == nil {
		for _, rb := range exactBooks {
			snippet := truncateStr(rb.Content, 200)
			result.RunbookSummaries = append(result.RunbookSummaries,
				fmt.Sprintf("[%s] %s", rb.Title, snippet))
		}
	}

	// ── 2. Milvus 语义召回 Runbook ─────────────────────────
	if query != "" {
		queryVec, embErr := aiEmbed.EmbedText(ctx, query)
		if embErr == nil {
			rbIDs, searchErr := milvusDAO.SearchRunbookSimilar(ctx, &settings.Conf.Milvus, queryVec, 3)
			if searchErr == nil {
				for _, id := range rbIDs {
					rb, dbErr := mysqlDAO.GetRunbookByID(id)
					if dbErr != nil || rb == nil {
						continue
					}
					snippet := truncateStr(rb.Content, 200)
					// 去重
					key := fmt.Sprintf("[%s] %s", rb.Title, snippet)
					if !contains(result.RunbookSummaries, key) {
						result.RunbookSummaries = append(result.RunbookSummaries, key)
					}
				}
			}
		}
	}

	// ── 3. 历史 ops_report 检索（最近 5 条相似报告）──────────
	reports, _, err := mysqlDAO.ListOpsReports(1, 5)
	if err == nil {
		for _, r := range reports {
			if strings.Contains(r.Summary, alertName) || strings.Contains(r.Summary, service) {
				result.ReportSummaries = append(result.ReportSummaries,
					fmt.Sprintf("[历史报告] %s", truncateStr(r.Summary, 150)))
			}
		}
	}

	return result, nil
}

// IndexRunbook 将 Runbook 内容写入 Milvus 向量索引
func IndexRunbook(ctx context.Context, runbookID int64, title, content string) error {
	chunks := splitRunbookContent(content, title, 400)
	if len(chunks) == 0 {
		return nil
	}

	embeddings, err := aiEmbed.EmbedBatch(ctx, chunks)
	if err != nil {
		return fmt.Errorf("index runbook: embed: %w", err)
	}

	rbChunks := make([]milvusDAO.RunbookChunkVector, len(chunks))
	for i, emb := range embeddings {
		rbChunks[i] = milvusDAO.RunbookChunkVector{
			ID:         snowflake.GenID(),
			RunbookID:  runbookID,
			ChunkIndex: int32(i),
			Embedding:  emb,
		}
	}

	return milvusDAO.InsertRunbookChunks(ctx, settings.Conf.Milvus.CollectionName, rbChunks)
}

// splitRunbookContent 将 Markdown 内容切分，标题注入每段方便语义召回
func splitRunbookContent(content, title string, maxLen int) []string {
	prefix := "[" + title + "] "
	paragraphs := strings.Split(content, "\n\n")
	var chunks []string
	var buf strings.Builder
	buf.WriteString(prefix)

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if buf.Len()+len([]rune(p)) > maxLen && buf.Len() > len(prefix) {
			chunks = append(chunks, buf.String())
			buf.Reset()
			buf.WriteString(prefix)
		}
		if buf.Len() > len(prefix) {
			buf.WriteString("\n\n")
		}
		buf.WriteString(p)
	}
	if buf.Len() > len(prefix) {
		chunks = append(chunks, buf.String())
	}
	return chunks
}

func truncateStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
