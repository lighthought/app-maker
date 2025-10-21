package config

import (
	"time"

	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port          string `mapstructure:"port"`
	Environment   string `mapstructure:"environment"`
	WorkspacePath string `mapstructure:"workspace_path"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type CommandConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
	CliTool string        `mapstructure:"cli_tool"`
}

// Asynq 异步配置
type AsynqConfig struct {
	Concurrency int `mapstructure:"concurrency"` // 并发数
}

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	Log     LogConfig     `mapstructure:"log"`
	Command CommandConfig `mapstructure:"command"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Asynq   AsynqConfig   `mapstructure:"asynq"`
}

// GitConfig Git配置
type GitConfig struct {
	UserID        string
	GUID          string
	ProjectPath   string
	CommitMessage string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("AGENTS")
	v.AutomaticEnv()

	v.SetDefault("app.port", "8088")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.workspace_path", "F:/app-maker/app_data")
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
