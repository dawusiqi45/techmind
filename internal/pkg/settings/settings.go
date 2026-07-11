package settings

import (
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
	RateLimit RateLimitSetting `mapstructure:"rateLimit"`
}

type AppSetting struct {
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"` // local / dev / prod
}

type ServerSetting struct {
	Addr string `mapstructure:"addr"`
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
	for _, key := range []string{
		"app.mode",
		"server.addr",
		"log.level",
		"mysql.dsn",
		"redis.addr",
		"redis.password",
		"milvus.addr",
		"jwt.secret",
		"ai.llmApiKey",
		"ai.embeddingApiKey",
		"monitor.prometheusURL",
		"alert.webhookToken",
	} {
		if err := v.BindEnv(key); err != nil {
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
	Conf.Log.Level = v.GetString("log.level")
	Conf.MySQL.DSN = v.GetString("mysql.dsn")
	Conf.Redis.Addr = v.GetString("redis.addr")
	Conf.Redis.Password = v.GetString("redis.password")
	Conf.Milvus.Addr = v.GetString("milvus.addr")
	Conf.JWT.Secret = v.GetString("jwt.secret")
	Conf.AI.LLMAPIKey = v.GetString("ai.llmApiKey")
	Conf.AI.EmbeddingAPIKey = v.GetString("ai.embeddingApiKey")
	Conf.Monitor.PrometheusURL = v.GetString("monitor.prometheusURL")
	Conf.Alert.WebhookToken = v.GetString("alert.webhookToken")
	return nil
}
