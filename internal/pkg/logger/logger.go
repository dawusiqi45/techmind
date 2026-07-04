package logger

import (
	"os"

	"techmind/internal/pkg/settings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init 根据配置初始化全局 Zap 日志器。
// 调用后全项目通过 zap.L() / zap.S() 使用。
func Init(cfg *settings.LogSetting, mode string) error {
	fileSync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   true,
	})

	jsonEncoderCfg := zap.NewProductionEncoderConfig()
	jsonEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoderCfg := zap.NewDevelopmentEncoderConfig()
	consoleEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	level := parseLevel(cfg.Level)

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(jsonEncoderCfg),
		fileSync,
		level,
	)

	var core zapcore.Core
	if mode == "local" {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderCfg),
			zapcore.AddSync(os.Stdout),
			level,
		)
		core = zapcore.NewTee(consoleCore, fileCore)
	} else {
		core = fileCore
	}

	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
	return nil
}

func parseLevel(s string) zapcore.Level {
	switch s {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
