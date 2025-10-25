package config

import (
	"time"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/spf13/viper"
)

// AppConfig App配置
type AppConfig struct {
	Port          string `mapstructure:"port"`           // 端口
	Environment   string `mapstructure:"environment"`    // 环境
	WorkspacePath string `mapstructure:"workspace_path"` // 工作空间路径
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"` // 日志级别
	File  string `mapstructure:"file"`  // 日志文件路径
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"` // 主机
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"` // 密码
	DB       int    `mapstructure:"db"`       // 数据库
}

// CommandConfig 命令配置
type CommandConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`  // 超时时间
	CliTool string        `mapstructure:"cli_tool"` // 命令行工具
}

// Asynq 异步配置
type AsynqConfig struct {
	Concurrency int `mapstructure:"concurrency"` // 并发数
}

// Config 配置
type Config struct {
	App     AppConfig     `mapstructure:"app"`     // App配置
	Log     LogConfig     `mapstructure:"log"`     // 日志配置
	Command CommandConfig `mapstructure:"command"` // 命令配置
	Redis   RedisConfig   `mapstructure:"redis"`   // Redis配置
	Asynq   AsynqConfig   `mapstructure:"asynq"`   // 异步配置
}

// GitConfig Git配置
type GitConfig struct {
	UserID        string // 用户ID
	GUID          string // 项目GUID
	ProjectPath   string // 项目路径
	CommitMessage string // 提交信息
}

// Load 加载配置
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("AGENTS")
	v.AutomaticEnv()

	v.SetDefault("app.port", "8088")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.workspace_path", utils.GetEnvOrDefault(common.EnvKeyWorkspacePath, utils.LOCAL_WORKSPACE_PATH))
	v.SetDefault("log.level", "debug")
	v.SetDefault("log.file", "./logs/app-maker-agents.log")
	v.SetDefault("command.timeout", "5m")
	v.SetDefault("command.cli_tool", "claude")
	v.SetDefault("redis.host", utils.GetEnvOrDefault("REDIS_HOST", "localhost"))
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", utils.GetEnvOrDefault("REDIS_PASSWORD", ""))
	v.SetDefault("redis.db", 1)
	v.SetDefault("asynq.concurrency", 100)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if cfg.Command.Timeout == 0 {
		cfg.Command.Timeout = 5 * time.Minute
	}

	return cfg, nil
}
