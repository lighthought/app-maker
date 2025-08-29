package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	BMad     BMadConfig     `mapstructure:"bmad"`
	Log      LogConfig      `mapstructure:"log"`
}

type AppConfig struct {
	Environment string `mapstructure:"environment"`
	Port        string `mapstructure:"port"`
	SecretKey   string `mapstructure:"secret_key"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Name           string `mapstructure:"name"`
	SSLMode        string `mapstructure:"ssl_mode"`
	ConnectTimeout int    `mapstructure:"connect_timeout"`
	AutoMigrate    bool   `mapstructure:"auto_migrate"`
	SeedData       bool   `mapstructure:"seed_data"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Expire    int    `mapstructure:"expire_hours"`
}

type BMadConfig struct {
	NpmPackage string `mapstructure:"npm_package"`
	ConfigPath string `mapstructure:"config_path"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")

	// 环境变量覆盖
	viper.AutomaticEnv()

	// 设置环境变量映射
	setEnvKeyReplacer()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用环境变量和默认值
		fmt.Printf("配置文件未找到，使用环境变量和默认值: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必要的配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// setEnvKeyReplacer 设置环境变量键名替换器
func setEnvKeyReplacer() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func setDefaults() {
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.secret_key", "your-secret-key-change-in-production")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5434)
	viper.SetDefault("database.user", "autocodeweb")
	viper.SetDefault("database.password", "AutoCodeWeb2024!@#")
	viper.SetDefault("database.name", "autocodeweb")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.connect_timeout", 10)
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.seed_data", true)

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("jwt.secret_key", "your-jwt-secret-key-change-in-production")
	viper.SetDefault("jwt.expire_hours", 24)

	viper.SetDefault("bmad.npm_package", "bmad-method")
	viper.SetDefault("bmad.config_path", "./bmad-config")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "./logs/app.log")
}

func validateConfig(config *Config) error {
	if config.App.Port == "" {
		return fmt.Errorf("应用端口不能为空")
	}

	if config.Database.Host == "" || config.Database.Name == "" {
		return fmt.Errorf("数据库配置不完整")
	}

	if config.JWT.SecretKey == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}

	return nil
}
