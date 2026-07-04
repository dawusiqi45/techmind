package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"techmind/internal/dao/mysql"
	"techmind/internal/dao/redis"
	"techmind/internal/dao/milvus"
	aiEmbed "techmind/internal/ai/embedding"
	aiModel "techmind/internal/ai/model"
	"techmind/internal/pkg/jwt"
	"techmind/internal/pkg/logger"
	"techmind/internal/monitor"
	"techmind/internal/pkg/settings"
	"techmind/internal/pkg/snowflake"
	"techmind/internal/router"

	"go.uber.org/zap"
)

func main() {
	// 支持 -config 参数指定配置文件路径，默认 config/config.yaml
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "server: fatal: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	// 1. 加载配置
	if err := settings.Init(configPath); err != nil {
		return fmt.Errorf("init settings: %w", err)
	}
	cfg := settings.Conf

	// 2. 初始化日志
	if err := logger.Init(&cfg.Log, cfg.App.Mode); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer zap.L().Sync() //nolint:errcheck
	zap.L().Info("TechMind server starting", zap.String("mode", cfg.App.Mode))

	// 3. 初始化 Snowflake
	if err := snowflake.Init(1); err != nil {
		return fmt.Errorf("init snowflake: %w", err)
	}

	// 4. 初始化 JWT
	jwt.Init(&cfg.JWT)
	monitor.RegisterMetrics()

	// 5. 初始化 MySQL
	if err := mysql.Init(&cfg.MySQL); err != nil {
		return fmt.Errorf("init mysql: %w", err)
	}
	defer mysql.Close()
	zap.L().Info("MySQL connected")

	// 6. 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	defer redis.Close()
	zap.L().Info("Redis connected")

	// 7. 初始化 Milvus（失败不阻断启动，降级为关键词搜索）
	if err := milvus.Init(&cfg.Milvus); err != nil {
		zap.L().Warn("Milvus init failed, semantic search disabled", zap.Error(err))
	} else {
		defer milvus.Close()
		zap.L().Info("Milvus connected")
	}

	// 8. 初始化 AI（LLM + Embedding，失败不阻断启动）
	if err := aiModel.InitLLM(&cfg.AI); err != nil {
		zap.L().Warn("LLM init failed, AI features disabled", zap.Error(err))
	}
	if err := aiEmbed.InitEmbedding(&cfg.AI); err != nil {
		zap.L().Warn("Embedding init failed, semantic search disabled", zap.Error(err))
	}

	// 9. 初始化路由
	engine := router.Setup(cfg.App.Mode)

	// 8. 启动 HTTP Server（优雅退出）
	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: engine,
	}

	go func() {
		zap.L().Info("HTTP server listening", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Error("ListenAndServe failed", zap.Error(err))
		}
	}()

	// 等待 SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	zap.L().Info("server stopped gracefully")
	return nil
}
