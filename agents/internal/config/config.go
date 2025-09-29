package config

import (
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port          string `mapstructure:"PORT"`
	Environment   string `mapstructure:"ENV"`
	WorkspacePath string `mapstructure:"WORKSPACE_PATH"`
}

type LogConfig struct {
	Level string `mapstructure:"LEVEL"`
	File  string `mapstructure:"FILE"`
}

type CommandConfig struct {
	Timeout          time.Duration `mapstructure:"TIMEOUT"`
	ClaudeBinaryPath string        `mapstructure:"CLAUDE_BIN"`
}

type Config struct {
	App     AppConfig     `mapstructure:",squash"`
	Log     LogConfig     `mapstructure:"LOG"`
	Command CommandConfig `mapstructure:"COMMAND"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("AGENTS")
	v.AutomaticEnv()

	v.SetDefault("PORT", "8088")
	v.SetDefault("ENV", "development")
	v.SetDefault("WORKSPACE_PATH", "F:/app-maker/app_data")
	v.SetDefault("LOG.LEVEL", "info")
	v.SetDefault("COMMAND.TIMEOUT", "5m")
	v.SetDefault("COMMAND.CLAUDE_BIN", "claude")

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
