package settings

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Conf 是全局配置对象，初始化后供全项目使用
var Conf = new(AppConfig)

// AppConfig 是整个应用的配置根结构
type AppConfig struct {
	App       AppSetting       `mapstructure:"app"`
	Server    ServerSetting    `mapstructure:"server"`
	Log       LogSetting       `mapstructure:"log"`
	MySQL     MySQLSetting     `mapstructure:"mysql"`
	Redis     RedisSetting     `mapstructure:"redis"`
	JWT       JWTSetting       `mapstructure:"jwt"`
	Milvus    MilvusSetting    `mapstructure:"milvus"`
	AI        AISetting        `mapstructure:"ai"`
	Monitor   MonitorSetting   `mapstructure:"monitor"`
	Alert     AlertSetting     `mapstructure:"alert"`
	Ops       OpsSetting       `mapstructure:"ops"`
	RateLimit RateLimitSetting `mapstructure:"rateLimit"`
}

type AppSetting struct {
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"` // local / dev / prod
}

type ServerSetting struct {
	Addr           string   `mapstructure:"addr"`
	TrustedProxies []string `mapstructure:"trustedProxies"`
}

type LogSetting struct {
	Level      string `mapstructure:"level"`      // debug / info / warn / error
	Filename   string `mapstructure:"filename"`   // logs/techmind.log
	MaxSize    int    `mapstructure:"maxSize"`    // MB
	MaxAge     int    `mapstructure:"maxAge"`     // days
	MaxBackups int    `mapstructure:"maxBackups"` // count
}

type MySQLSetting struct {
	DSN         string `mapstructure:"dsn"`
	MaxOpen     int    `mapstructure:"maxOpen"`
	MaxIdle     int    `mapstructure:"maxIdle"`
	MaxLifetime int    `mapstructure:"maxLifetime"` // seconds
}

type RedisSetting struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTSetting struct {
	Secret          string `mapstructure:"secret"`
	AccessExpireMin int    `mapstructure:"accessExpireMin"` // minutes
	RefreshExpireH  int    `mapstructure:"refreshExpireH"`  // hours
}

type MilvusSetting struct {
	Addr           string `mapstructure:"addr"`           // e.g. localhost:19530
	CollectionName string `mapstructure:"collectionName"` // e.g. techmind_articles
	Dim            int    `mapstructure:"dim"`            // embedding dimension, e.g. 1536
}

// AISetting 聚合 LLM 和 Embedding 配置
type AISetting struct {
	// LLM: OpenAI 兼容接口（支持 DeepSeek / Doubao）
	LLMBaseURL string `mapstructure:"llmBaseURL"` // e.g. https://api.deepseek.com/v1
	LLMAPIKey  string `mapstructure:"llmApiKey"`
	LLMModel   string `mapstructure:"llmModel"` // e.g. deepseek-chat

	// Embedding: 字节跳动 Ark（DashScope text-embedding-v4 兼容）
	EmbeddingAPIKey string `mapstructure:"embeddingApiKey"`
	EmbeddingModel  string `mapstructure:"embeddingModel"` // e.g. ep-xxx (Ark endpoint)

	// 超时和限流
	TimeoutSec     int `mapstructure:"timeoutSec"`     // default 30
	MaxConcurrency int `mapstructure:"maxConcurrency"` // default 5
}

// MonitorSetting 定义 Agent 查询 Prometheus 所需的地址。
type MonitorSetting struct {
	PrometheusURL string `mapstructure:"prometheusURL"`
}

// AlertSetting 定义 Alertmanager Webhook 的共享令牌。
type AlertSetting struct {
	WebhookToken string `mapstructure:"webhookToken"`
}

// OpsSetting 定义 SRE Agent 的安全边界和自动触发策略。
type OpsSetting struct {
	AutoDiagnose       bool `mapstructure:"autoDiagnose"`
	DiagnoseTimeoutSec int  `mapstructure:"diagnoseTimeoutSec"`
	EvidenceWindowMin  int  `mapstructure:"evidenceWindowMin"`
}

// RateLimitSetting 限流配置
type RateLimitSetting struct {
	Enabled        bool `mapstructure:"enabled"`
	RequestsPerMin int  `mapstructure:"requestsPerMin"` // 每分钟允许请求数，默认 60
}

// Init 从指定路径加载配置文件（YAML），解码到 Conf。
// configPath 例如 "config/config.yaml"
func Init(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("TECHMIND")
	v.AutomaticEnv()
	// 显式绑定环境变量。Viper 仅把 '.' 替换为 '_'，不会自动在驼峰字段
	// 内插入下划线；因此 AI/Webhook 等字段不能只依赖默认 BindEnv 行为。
	envBindings := map[string]string{
		"app.mode":               "TECHMIND_APP_MODE",
		"server.addr":            "TECHMIND_SERVER_ADDR",
		"server.trustedProxies":  "TECHMIND_SERVER_TRUSTED_PROXIES",
		"log.level":              "TECHMIND_LOG_LEVEL",
		"mysql.dsn":              "TECHMIND_MYSQL_DSN",
		"redis.addr":             "TECHMIND_REDIS_ADDR",
		"redis.password":         "TECHMIND_REDIS_PASSWORD",
		"milvus.addr":            "TECHMIND_MILVUS_ADDR",
		"jwt.secret":             "TECHMIND_JWT_SECRET",
		"ai.llmBaseURL":          "TECHMIND_AI_LLM_BASE_URL",
		"ai.llmApiKey":           "TECHMIND_AI_LLM_API_KEY",
		"ai.llmModel":            "TECHMIND_AI_LLM_MODEL",
		"ai.embeddingApiKey":     "TECHMIND_AI_EMBEDDING_API_KEY",
		"ai.embeddingModel":      "TECHMIND_AI_EMBEDDING_MODEL",
		"monitor.prometheusURL":  "TECHMIND_MONITOR_PROMETHEUS_URL",
		"alert.webhookToken":     "TECHMIND_ALERT_WEBHOOK_TOKEN",
		"ops.autoDiagnose":       "TECHMIND_OPS_AUTO_DIAGNOSE",
		"ops.diagnoseTimeoutSec": "TECHMIND_OPS_DIAGNOSE_TIMEOUT_SEC",
		"ops.evidenceWindowMin":  "TECHMIND_OPS_EVIDENCE_WINDOW_MIN",
	}
	for key, envName := range envBindings {
		if err := v.BindEnv(key, envName); err != nil {
			return fmt.Errorf("settings: bind env %q: %w", key, err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("settings: read config file %q failed: %w", configPath, err)
	}
	if err := v.Unmarshal(Conf); err != nil {
		return fmt.Errorf("settings: unmarshal config failed: %w", err)
	}
	// viper 的 Unmarshal 不保证把环境变量覆盖写回嵌套结构，显式同步所有
	// 支持环境变量的运行时字段，确保 Docker/Helm 注入一定生效。
	Conf.App.Mode = v.GetString("app.mode")
	Conf.Server.Addr = v.GetString("server.addr")
	Conf.Server.TrustedProxies = v.GetStringSlice("server.trustedProxies")
	Conf.Log.Level = v.GetString("log.level")
	Conf.MySQL.DSN = v.GetString("mysql.dsn")
	Conf.Redis.Addr = v.GetString("redis.addr")
	Conf.Redis.Password = v.GetString("redis.password")
	Conf.Milvus.Addr = v.GetString("milvus.addr")
	Conf.JWT.Secret = v.GetString("jwt.secret")
	Conf.AI.LLMBaseURL = v.GetString("ai.llmBaseURL")
	Conf.AI.LLMAPIKey = v.GetString("ai.llmApiKey")
	Conf.AI.LLMModel = v.GetString("ai.llmModel")
	Conf.AI.EmbeddingAPIKey = v.GetString("ai.embeddingApiKey")
	Conf.AI.EmbeddingModel = v.GetString("ai.embeddingModel")
	Conf.Monitor.PrometheusURL = v.GetString("monitor.prometheusURL")
	Conf.Alert.WebhookToken = v.GetString("alert.webhookToken")
	Conf.Ops.AutoDiagnose = v.GetBool("ops.autoDiagnose")
	Conf.Ops.DiagnoseTimeoutSec = v.GetInt("ops.diagnoseTimeoutSec")
	Conf.Ops.EvidenceWindowMin = v.GetInt("ops.evidenceWindowMin")
	weakJWTSecret := len(Conf.JWT.Secret) < 32 || Conf.JWT.Secret == "change-me-in-production"
	if Conf.App.Mode != "local" && weakJWTSecret {
		return fmt.Errorf("settings: jwt secret must be at least 32 characters and not use the default outside local mode")
	}
	if Conf.App.Mode == "local" && weakJWTSecret {
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return fmt.Errorf("settings: generate local jwt secret: %w", err)
		}
		Conf.JWT.Secret = hex.EncodeToString(secret)
	}
	return nil
}
