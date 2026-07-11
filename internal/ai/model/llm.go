package ai

import (
	"context"
	"fmt"
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/monitor"
	"techmind/internal/pkg/settings"
	"techmind/internal/pkg/snowflake"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	einoModel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// llmClient 全局 LLM 客户端
var llmClient einoModel.ChatModel

// llmModelName 当前使用的模型名称（用于写 ai_call_record）
var llmModelName string

// InitLLM 初始化 OpenAI 兼容 LLM 客户端（支持 DeepSeek/Doubao）
func InitLLM(cfg *settings.AISetting) error {
	if cfg.LLMBaseURL == "" || cfg.LLMAPIKey == "" || cfg.LLMModel == "" {
		return fmt.Errorf("ai: llm configuration is incomplete")
	}
	c, err := einoOpenAI.NewChatModel(context.Background(), &einoOpenAI.ChatModelConfig{
		BaseURL: cfg.LLMBaseURL,
		APIKey:  cfg.LLMAPIKey,
		Model:   cfg.LLMModel,
		Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("ai: init llm: %w", err)
	}
	llmClient = c
	llmModelName = cfg.LLMModel
	return nil
}

// Chat 调用 LLM，返回第一条回复文本，并按 Skill 写入调用审计。
func Chat(ctx context.Context, skill, systemPrompt, userMsg string) (string, error) {
	if llmClient == nil {
		return "", fmt.Errorf("ai: llm client not initialized")
	}
	msgs := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMsg),
	}
	start := time.Now()
	resp, err := llmClient.Generate(ctx, msgs)
	duration := time.Since(start)
	monitor.ObserveAICall(skill, duration, err)

	// 异步写 ai_call_record
	go writeCallRecord(ctx, skill, llmModelName, int(duration.Milliseconds()), err)

	if err != nil {
		return "", fmt.Errorf("ai: llm generate: %w", err)
	}
	return resp.Content, nil
}

// ChatOnce 为既有调用保留兼容入口。
func ChatOnce(ctx context.Context, systemPrompt, userMsg string) (string, error) {
	return Chat(ctx, "llm", systemPrompt, userMsg)
}

// writeCallRecord 异步写入 AI 调用记录，失败只记日志不影响主流程
func writeCallRecord(ctx context.Context, skill, modelName string, durationMs int, callErr error) {
	status := "ok"
	errMsg := ""
	if callErr != nil {
		status = "failed"
		errMsg = callErr.Error()
		if len(errMsg) > 255 {
			errMsg = errMsg[:255]
		}
	}
	r := &model.AICallRecord{
		ID:         snowflake.GenID(),
		Skill:      skill,
		Model:      modelName,
		DurationMs: durationMs,
		Status:     status,
		ErrorMsg:   errMsg,
	}
	_ = mysqlDAO.CreateAICallRecord(r)
}
