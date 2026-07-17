package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitReadsStandardSecretEnvironmentNames(t *testing.T) {
	t.Setenv("TECHMIND_AI_LLM_API_KEY", "llm-from-env")
	t.Setenv("TECHMIND_AI_LLM_BASE_URL", "https://llm.example/v1")
	t.Setenv("TECHMIND_AI_LLM_MODEL", "example-chat")
	t.Setenv("TECHMIND_AI_EMBEDDING_API_KEY", "embedding-from-env")
	t.Setenv("TECHMIND_AI_EMBEDDING_MODEL", "example-embedding")
	t.Setenv("TECHMIND_ALERT_WEBHOOK_TOKEN", "webhook-from-env")
	t.Setenv("TECHMIND_OPS_AUTO_DIAGNOSE", "true")
	t.Setenv("TECHMIND_OPS_DIAGNOSE_TIMEOUT_SEC", "90")
	t.Setenv("TECHMIND_OPS_EVIDENCE_WINDOW_MIN", "20")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	config := "app:\n  mode: local\nai:\n  llmBaseURL: ''\n  llmApiKey: ''\n  llmModel: ''\n  embeddingApiKey: ''\n  embeddingModel: ''\nalert:\n  webhookToken: ''\nops:\n  autoDiagnose: false\n  diagnoseTimeoutSec: 120\n  evidenceWindowMin: 30\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	Conf = new(AppConfig)
	if err := Init(configPath); err != nil {
		t.Fatalf("init settings: %v", err)
	}
	if Conf.AI.LLMAPIKey != "llm-from-env" {
		t.Fatalf("LLM key was not read from standard env name: %q", Conf.AI.LLMAPIKey)
	}
	if Conf.AI.LLMBaseURL != "https://llm.example/v1" || Conf.AI.LLMModel != "example-chat" {
		t.Fatalf("LLM provider settings were not read from standard env names: %#v", Conf.AI)
	}
	if Conf.AI.EmbeddingAPIKey != "embedding-from-env" {
		t.Fatalf("embedding key was not read from standard env name: %q", Conf.AI.EmbeddingAPIKey)
	}
	if Conf.AI.EmbeddingModel != "example-embedding" {
		t.Fatalf("embedding model was not read from standard env name: %q", Conf.AI.EmbeddingModel)
	}
	if Conf.Alert.WebhookToken != "webhook-from-env" {
		t.Fatalf("webhook token was not read from standard env name: %q", Conf.Alert.WebhookToken)
	}
	if !Conf.Ops.AutoDiagnose || Conf.Ops.DiagnoseTimeoutSec != 90 || Conf.Ops.EvidenceWindowMin != 20 {
		t.Fatalf("ops settings were not read from standard env names: %#v", Conf.Ops)
	}
}
