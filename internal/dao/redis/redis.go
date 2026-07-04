package redis

import (
	"context"
	"fmt"

	"techmind/internal/pkg/settings"

	goredis "github.com/redis/go-redis/v9"
)

// RDB 是全局 Redis 客户端，初始化后供业务层使用
var RDB *goredis.Client

// Init 根据配置初始化 Redis 客户端并验证连通性
func Init(cfg *settings.RedisSetting) error {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return fmt.Errorf("redis: ping failed: %w", err)
	}

	RDB = rdb
	return nil
}

// Close 关闭 Redis 连接，应在进程退出前调用
func Close() {
	if RDB != nil {
		_ = RDB.Close()
	}
}
