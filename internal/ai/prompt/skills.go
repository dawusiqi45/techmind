// Package prompt 实现 TechMind 的 AI Skill 层，封装摘要、标签生成和搜索总结能力。
// 每个 Skill 是一个无状态函数，接收业务数据，返回 AI 生成结果。
// 依赖 internal/ai/model.llmClient，调用前必须 InitLLM。
package prompt

import (
	"context"
	"fmt"
	"strings"

	llm "techmind/internal/ai/model"
)

// ArticleSummarySkill 根据文章标题和正文生成摘要（≤200字）
func ArticleSummarySkill(ctx context.Context, title, content string) (string, error) {
	system := "你是一个技术内容摘要助手。请根据文章标题和正文，生成一段简洁的中文摘要，不超过200字，不要使用markdown格式。"
	user := fmt.Sprintf("标题：%s\n\n正文（节选）：%s", title, truncate(content, 2000))
	summary, err := llm.ChatOnce(ctx, system, user)
	if err != nil {
		return "", fmt.Errorf("ArticleSummarySkill: %w", err)
	}
	return strings.TrimSpace(summary), nil
}

// ArticleTagSkill 根据文章标题和正文生成 AI 标签（3-5个）
// 返回标签名列表
func ArticleTagSkill(ctx context.Context, title, content string) ([]string, error) {
	system := "你是一个技术文章分类助手。请根据文章内容，提取3到5个技术标签，用英文逗号分隔，只输出标签不输出其他内容，例如：Go,微服务,Kubernetes"
	user := fmt.Sprintf("标题：%s\n\n正文（节选）：%s", title, truncate(content, 1500))
	result, err := llm.ChatOnce(ctx, system, user)
	if err != nil {
		return nil, fmt.Errorf("ArticleTagSkill: %w", err)
	}
	tags := splitTags(result)
	return tags, nil
}

// SearchSummarySkill 根据搜索关键词和 TopN 文章摘要生成搜索总结
func SearchSummarySkill(ctx context.Context, keyword string, summaries []string) (string, error) {
	if len(summaries) == 0 {
		return "", nil
	}
	system := "你是一个搜索结果总结助手。请根据搜索关键词和相关文章摘要，用2-3句话总结这批搜索结果的主要内容，用中文，不超过150字。"
	user := fmt.Sprintf("搜索词：%s\n\n相关文章摘要：\n%s", keyword, strings.Join(summaries, "\n"))
	summary, err := llm.ChatOnce(ctx, system, user)
	if err != nil {
		return "", fmt.Errorf("SearchSummarySkill: %w", err)
	}
	return strings.TrimSpace(summary), nil
}

// truncate 截断文本到 maxLen 字符
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// splitTags 将逗号分隔的标签字符串拆分并清理
func splitTags(s string) []string {
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
