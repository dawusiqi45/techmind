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
	system := `你是 TechMind 的技术内容摘要 Skill。只能依据提供的文章内容总结，不能补充不存在的事实。
要求：中文、不超过 200 字、不使用 Markdown；覆盖问题、关键技术和结论；若正文信息不足，明确说明信息不足。`
	user := fmt.Sprintf("标题：%s\n\n正文（节选）：%s", title, truncate(content, 2000))
	summary, err := llm.Chat(ctx, "article_summary", system, user)
	if err != nil {
		return "", fmt.Errorf("ArticleSummarySkill: %w", err)
	}
	return strings.TrimSpace(summary), nil
}

// ArticleTagSkill 根据文章标题和正文生成 AI 标签（3-5个）
// 返回标签名列表
func ArticleTagSkill(ctx context.Context, title, content string) ([]string, error) {
	system := `你是 TechMind 的文章分类 Skill。只能根据文章内容生成 3 到 5 个准确、可检索的技术标签。
标签应为具体技术或领域，不要使用“技术”“文章”等泛词；使用英文逗号分隔；只输出标签，不输出解释或 Markdown。`
	user := fmt.Sprintf("标题：%s\n\n正文（节选）：%s", title, truncate(content, 1500))
	result, err := llm.Chat(ctx, "article_tag", system, user)
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
	system := `你是 TechMind 的搜索总结 Skill。只能根据给出的搜索词和文章摘要归纳结果，不得虚构文章内容。
用中文 2 到 3 句话、最多 150 字；优先说明共同主题、差异和适用场景；证据不足时明确说明。`
	user := fmt.Sprintf("搜索词：%s\n\n相关文章摘要：\n%s", keyword, strings.Join(summaries, "\n"))
	summary, err := llm.Chat(ctx, "search_summary", system, user)
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
