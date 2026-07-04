package health

import (
	"context"
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	milvusDAO "techmind/internal/dao/milvus"
)

// CheckResult 单项检查结果
type CheckResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// ReadyzResult 就绪检查汇总
type ReadyzResult struct {
	Ready  bool                    `json:"ready"`
	Checks map[string]CheckResult  `json:"checks"`
}

// Readyz 检查 MySQL、Redis、Milvus 连通性，任意一项失败则 ready=false
func Readyz() *ReadyzResult {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)

	// MySQL
	if mysqlDAO.DB != nil {
		sqlDB, err := mysqlDAO.DB.DB()
		if err == nil {
			err = sqlDB.PingContext(ctx)
		}
		if err != nil {
			checks["mysql"] = CheckResult{OK: false, Message: err.Error()}
		} else {
			checks["mysql"] = CheckResult{OK: true}
		}
	} else {
		checks["mysql"] = CheckResult{OK: false, Message: "not initialized"}
	}

	// Redis
	if redisDAO.RDB != nil {
		if err := redisDAO.RDB.Ping(ctx).Err(); err != nil {
			checks["redis"] = CheckResult{OK: false, Message: err.Error()}
		} else {
			checks["redis"] = CheckResult{OK: true}
		}
	} else {
		checks["redis"] = CheckResult{OK: false, Message: "not initialized"}
	}

	// Milvus（可选：未初始化视为 degraded 但不影响 ready）
	if milvusDAO.Client != nil {
		_, err := milvusDAO.Client.ListCollections(ctx)
		if err != nil {
			checks["milvus"] = CheckResult{OK: false, Message: err.Error()}
		} else {
			checks["milvus"] = CheckResult{OK: true}
		}
	} else {
		checks["milvus"] = CheckResult{OK: false, Message: "not initialized (degraded)"}
	}

	// 整体 ready 条件：MySQL + Redis 必须 OK，Milvus 允许降级
	ready := checks["mysql"].OK && checks["redis"].OK

	return &ReadyzResult{
		Ready:  ready,
		Checks: checks,
	}
}
