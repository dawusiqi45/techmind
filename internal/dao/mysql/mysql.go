package mysql

import (
	"fmt"
	"time"

	"techmind/internal/pkg/settings"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局 GORM 数据库实例，初始化后供 DAO 层使用
var DB *gorm.DB

// Init 根据配置初始化 GORM MySQL 连接池并验证连通性
func Init(cfg *settings.MySQLSetting) error {
	gormCfg := &gorm.Config{
		// 生产环境建议设置为 logger.Silent，这里开发阶段用 Warn
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(gormmysql.Open(cfg.DSN), gormCfg)
	if err != nil {
		return fmt.Errorf("mysql: gorm open failed: %w", err)
	}

	// 获取底层 *sql.DB 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("mysql: get underlying sql.DB failed: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// 验证连通性
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("mysql: ping failed: %w", err)
	}

	DB = db
	return nil
}

// Close 关闭连接池，应在进程退出前调用
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}
