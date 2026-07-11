package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	aiEmbed "techmind/internal/ai/embedding"
	aiModel "techmind/internal/ai/model"
	"techmind/internal/dao/milvus"
	"techmind/internal/dao/mysql"
	"techmind/internal/dao/redis"
	"techmind/internal/monitor"
	"techmind/internal/pkg/logger"
	"techmind/internal/pkg/settings"
	"techmind/internal/pkg/snowflake"
	"techmind/internal/worker"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "worker: fatal: %v\n", err)
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
	zap.L().Info("TechMind worker starting", zap.String("mode", cfg.App.Mode))

	// 3. 初始化 Snowflake（worker 使用 machineID=2，与 server 区分）
	if err := snowflake.Init(2); err != nil {
		return fmt.Errorf("init snowflake: %w", err)
	}
	monitor.RegisterMetrics()

	// 4. 初始化 MySQL
	if err := mysql.Init(&cfg.MySQL); err != nil {
		return fmt.Errorf("init mysql: %w", err)
	}
	defer mysql.Close()
	zap.L().Info("MySQL connected")

	// 5. 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	defer redis.Close()
	zap.L().Info("Redis connected")

	// 6. 初始化 Milvus
	if err := milvus.Init(&cfg.Milvus); err != nil {
		zap.L().Warn("Milvus init failed, index tasks will fail", zap.Error(err))
	} else {
		defer milvus.Close()
		zap.L().Info("Milvus connected")
	}

	// 7. 初始化 AI（LLM + Embedding）
	if err := aiModel.InitLLM(&cfg.AI); err != nil {
		zap.L().Warn("LLM init failed", zap.Error(err))
	}
	if err := aiEmbed.InitEmbedding(&cfg.AI); err != nil {
		zap.L().Warn("Embedding init failed", zap.Error(err))
	}

	// 8. 创建 AI Worker 并注册处理器
	aiWorker := worker.NewAIWorker("ai-worker-1")
	worker.RegisterAIHandlers(aiWorker)

	// 创建 Ops Worker（诊断任务）
	opsWorker := worker.NewOpsWorker("ops-worker-1")
	metricsServer := &http.Server{Addr: ":9091", Handler: promhttp.Handler()}
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Warn("worker metrics server stopped", zap.Error(err))
		}
	}()

	// 9. 启动 Worker（可取消的 context）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = metricsServer.Shutdown(shutdownCtx)
	}()

	errCh := make(chan error, 2)
	go func() {
		errCh <- aiWorker.Start(ctx)
	}()
	go func() {
		errCh <- opsWorker.Start(ctx)
	}()

	// 等待 SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		zap.L().Info("received signal, shutting down worker", zap.String("signal", sig.String()))
		cancel()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("worker exited with error: %w", err)
		}
	}

	// 等待 worker goroutine 退出
	<-errCh
	zap.L().Info("worker stopped gracefully")
	return nil
}
