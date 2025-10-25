package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/spf13/viper"
)

// Config 配置
type Config struct {
	App      AppConfig      `mapstructure:"app"`      // App配置
	Database DatabaseConfig `mapstructure:"database"` // 数据库配置
	Redis    RedisConfig    `mapstructure:"redis"`    // Redis配置
	Asynq    AsynqConfig    `mapstructure:"asynq"`    // 异步配置
	CORS     CORSConfig     `mapstructure:"cors"`     // CORS配置
	JWT      JWTConfig      `mapstructure:"jwt"`      // JWT配置
	Log      LogConfig      `mapstructure:"log"`      // 日志配置
	Agents   AgentsConfig   `mapstructure:"agents"`   // Agents配置
}

// AppConfig App配置
type AppConfig struct {
	Environment string `mapstructure:"environment"` // 环境
	Port        string `mapstructure:"port"`        // 端口
	SecretKey   string `mapstructure:"secret_key"`  // 密钥
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host           string `mapstructure:"host"`            // 主机
	Port           int    `mapstructure:"port"`            // 端口
	User           string `mapstructure:"user"`            // 用户
	Password       string `mapstructure:"password"`        // 密码
	Name           string `mapstructure:"name"`            // 数据库名称
	SSLMode        string `mapstructure:"ssl_mode"`        // SSL模式
	ConnectTimeout int    `mapstructure:"connect_timeout"` // 连接超时时间
	AutoMigrate    bool   `mapstructure:"auto_migrate"`    // 自动迁移
	SeedData       bool   `mapstructure:"seed_data"`       // 种子数据
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`     // 主机
	Port     int    `mapstructure:"port"`     // 端口
	Password string `mapstructure:"password"` // 密码
	DB       int    `mapstructure:"db"`       // 数据库
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`   // 密钥
	Expire    int    `mapstructure:"expire_hours"` // 过期时间
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`   // 允许的源
	AllowedMethods   []string `mapstructure:"allowed_methods"`   // 允许的方法
	AllowedHeaders   []string `mapstructure:"allowed_headers"`   // 允许的头
	AllowCredentials bool     `mapstructure:"allow_credentials"` // 允许的凭据
	MaxAge           int      `mapstructure:"max_age"`           // 最大年龄
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"` // 日志级别
	File  string `mapstructure:"file"`  // 日志文件路径
}

// Agents server配置
type AgentsConfig struct {
	URL string `mapstructure:"url"` // Agents server URL
}

// Asynq 异步配置
type AsynqConfig struct {
	Concurrency int `mapstructure:"concurrency"` // 并发数
}

func Load() (*Config, error) {
	env := utils.GetEnvOrDefault("APP_ENVIRONMENT", "")
	switch env {
	case common.EnvironmentLocalDebug:
		viper.SetConfigName("config.local")
	case common.EnvironmentDevelopment:
		viper.SetConfigName("config")
	case common.EnvironmentProduction:
		viper.SetConfigName("config.prod")
	default:
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AutomaticEnv() // 环境变量覆盖
	setEnvKeyReplacer()  // 设置环境变量映射
	setDefaults()        // 设置默认值

	if err := viper.ReadInConfig(); err != nil { // 读取配置文件
		// 如果配置文件不存在，使用环境变量和默认值
		fmt.Printf("config file not found, using environment variables and defaults: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err.Error())
	}

	// 验证必要的配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("failed to validate config: %s", err.Error())
	}

	return &config, nil
}

// setEnvKeyReplacer 设置环境变量键名替换器
func setEnvKeyReplacer() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func setDefaults() {
	viper.SetDefault("app.environment", utils.GetEnvOrDefault("APP_ENVIRONMENT", "development"))
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.secret_key", utils.GetEnvOrDefault("APP_SECRET_KEY", "your-secret-key-change-in-production"))

	viper.SetDefault("database.host", utils.GetEnvOrDefault("DATABASE_HOST", "postgres"))
	viper.SetDefault("database.port", 5434)
	viper.SetDefault("database.user", utils.GetEnvOrDefault("DATABASE_USER", "autocodeweb"))
	viper.SetDefault("database.password", utils.GetEnvOrDefault("DATABASE_PASSWORD", "your-secret-key-change-in-production"))
	viper.SetDefault("database.name", utils.GetEnvOrDefault("DATABASE_NAME", "autocodeweb"))
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.connect_timeout", 10)
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.seed_data", true)

	viper.SetDefault("redis.host", utils.GetEnvOrDefault("REDIS_HOST", "redis"))
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", utils.GetEnvOrDefault("REDIS_PASSWORD", "your-secret-key-change-in-production"))
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("jwt.secret_key", utils.GetEnvOrDefault("JWT_SECRET_KEY", "your-jwt-secret-key-change-in-production"))
	viper.SetDefault("jwt.expire_hours", 24)

	viper.SetDefault("bmad.npm_package", "bmad-method")

	viper.SetDefault("ai.ollama.base_url", utils.GetEnvOrDefault("OLLAMA_URL", "http://chat.app-maker.localhost:11434"))
	viper.SetDefault("ai.ollama.model", utils.GetEnvOrDefault("OLLAMA_MODEL", "deepseek-r1:14b"))
	timeout, err := strconv.Atoi(utils.GetEnvOrDefault("OLLAMA_TIMEOUT", "60"))
	if err != nil {
		timeout = 60
	}
	viper.SetDefault("ai.ollama.timeout", timeout)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "./logs/app.log")

	// CORS 默认配置
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"*"})
	viper.SetDefault("cors.allow_credentials", false)
	viper.SetDefault("cors.max_age", 86400)

	viper.SetDefault("asynq.concurrency", 100)

	// Agents Server 默认
	viper.SetDefault("agents.url", utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://localhost:8088"))
}

func validateConfig(config *Config) error {
	if config.App.Port == "" {
		return fmt.Errorf("app port cannot be empty")
	}

	if config.Database.Host == "" || config.Database.Name == "" {
		return fmt.Errorf("database config is incomplete")
	}

	if config.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key cannot be empty")
	}

	return nil
}
