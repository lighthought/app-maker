package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Init 初始化日志系统
func Init(level, filePath string) error {
	// 解析日志级别
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	// 创建日志目录
	if filePath != "" {
		logDir := filepath.Dir(filePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
	}

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建核心
	var core zapcore.Core

	if filePath != "" {
		// 文件输出
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			logLevel,
		)

		// 控制台输出
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			logLevel,
		)

		// 组合多个核心
		core = zapcore.NewTee(fileCore, consoleCore)
	} else {
		// 仅控制台输出
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			logLevel,
		)
	}

	// 创建日志记录器
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Sync 同步日志
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// String 创建字符串字段
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int 创建整数字段
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Error 创建错误字段
func ErrorField(err error) zap.Field {
	return zap.Error(err)
}
