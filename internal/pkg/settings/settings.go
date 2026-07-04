package settings

import (
	"fmt"

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
	AccessExpireMin int    `mapstructure:"accessExpireMin"`  // minutes
	RefreshExpireH  int    `mapstructure:"refreshExpireH"`   // hours
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

// RateLimitSetting 限流配置
type RateLimitSetting struct {
	Enabled      bool `mapstructure:"enabled"`
	RequestsPerMin int `mapstructure:"requestsPerMin"` // 每分钟允许请求数，默认 60
}

// Init 从指定路径加载配置文件（YAML），解码到 Conf。
// configPath 例如 "config/config.yaml"
func Init(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv() // 允许环境变量覆盖

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("settings: read config file %q failed: %w", configPath, err)
	}
	if err := v.Unmarshal(Conf); err != nil {
		return fmt.Errorf("settings: unmarshal config failed: %w", err)
	}
	return nil
}
