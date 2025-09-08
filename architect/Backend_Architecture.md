# AutoCodeWeb 后端架构设计

## 1. 后端架构概述

### 1.1 架构理念
- **微服务架构**：按业务域划分服务，支持独立部署和扩展
- **分层架构**：清晰的层次分离，便于维护和测试
- **事件驱动**：基于消息队列的异步处理机制
- **RESTful API**：标准化的API设计，易于集成和扩展
- **高并发支持**：Go语言的协程特性，支持高并发处理

### 1.2 技术栈选型
- **编程语言**：Go 1.21+
- **Web框架**：Gin 1.9+
- **ORM框架**：GORM 1.25+
- **数据库**：PostgreSQL 15+
- **缓存**：Redis 7+
- **消息队列**：Redis Streams
- **配置管理**：Viper
- **日志系统**：Zap
- **验证框架**：validator
- **JWT认证**：golang-jwt

## 2. 项目结构设计

### 2.1 目录结构
```
backend/
├── cmd/                    # 应用程序入口
│   └── server/            # 主服务入口
│       └── main.go
├── internal/               # 内部包
│   ├── api/               # API层
│   │   ├── handlers/      # HTTP处理器
│   │   ├── middleware/    # 中间件
│   │   └── routes/        # 路由定义
│   ├── config/            # 配置管理
│   ├── database/          # 数据库相关
│   │   ├── migrations/    # 数据库迁移
│   │   └── seeds/         # 初始数据
│   ├── models/            # 数据模型
│   ├── repositories/      # 数据访问层
│   ├── services/          # 业务逻辑层
│   ├── utils/             # 工具函数
│   └── worker/            # 后台工作进程
├── pkg/                    # 可导出的包
│   ├── auth/              # 认证相关
│   ├── bmad/              # BMad-Method集成
│   ├── cache/             # 缓存管理
│   ├── logger/            # 日志管理
│   └── validator/         # 验证器
├── scripts/                # 脚本文件
├── .env                    # 环境变量
├── .env.example           # 环境变量示例
├── go.mod                 # Go模块文件
├── go.sum                 # 依赖校验文件
├── Dockerfile             # Docker构建文件
└── docker-compose.yml     # Docker编排文件
```

### 2.2 分层架构设计
```
┌─────────────────────────────────────────────────────────────┐
│                        API层 (API Layer)                    │
│                   HTTP处理器、路由、中间件                    │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      服务层 (Service Layer)                  │
│                  业务逻辑、事务管理、业务规则                  │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┘
│                    仓储层 (Repository Layer)                 │
│                  数据访问、查询优化、缓存管理                  │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      数据层 (Data Layer)                    │
│                  数据库、缓存、文件存储                      │
└─────────────────────────────────────────────────────────────┘
```

## 3. 核心架构设计

### 3.1 应用启动架构
```go
// cmd/server/main.go - 主程序入口
package main

import (
    "log"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/your-org/autocodeweb/internal/api/routes"
    "github.com/your-org/autocodeweb/internal/config"
    "github.com/your-org/autocodeweb/internal/database"
    "github.com/your-org/autocodeweb/pkg/logger"
)

func main() {
    // 加载配置
    cfg := config.Load()
    
    // 初始化日志
    logger := logger.New(cfg.Log)
    defer logger.Sync()
    
    // 连接数据库
    db, err := database.Connect(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // 连接Redis
    redis, err := database.ConnectRedis(cfg.Redis)
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer redis.Close()
    
    // 设置Gin模式
    if cfg.App.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    // 创建Gin引擎
    engine := gin.New()
    
    // 注册中间件
    engine.Use(gin.Logger(), gin.Recovery())
    
    // 注册路由
    routes.Register(engine, db, redis, cfg)
    
    // 启动服务器
    logger.Info("Server starting on port " + cfg.App.Port)
    if err := http.ListenAndServe(":"+cfg.App.Port, engine); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

### 3.2 配置管理架构
```go
// internal/config/config.go - 配置管理
package config

import (
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
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    Name     string `mapstructure:"name"`
    SSLMode  string `mapstructure:"ssl_mode"`
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

func Load() *Config {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    
    // 环境变量覆盖
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        panic("Failed to read config file: " + err.Error())
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        panic("Failed to unmarshal config: " + err.Error())
    }
    
    return &config
}
```

### 3.3 数据库连接架构
```go
// internal/database/connection.go - 数据库连接
package database

import (
    "fmt"
    "log"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    
    "github.com/your-org/autocodeweb/internal/config"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // 获取底层sql.DB对象
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB: %w", err)
    }
    
    // 设置连接池参数
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(3600) // 1小时
    
    log.Println("Database connected successfully")
    return db, nil
}

func ConnectRedis(cfg config.RedisConfig) (*redis.Client, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })
    
    // 测试连接
    if err := rdb.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }
    
    log.Println("Redis connected successfully")
    return rdb, nil
}
```

## 4. 服务层架构设计

### 4.1 用户服务架构
```go
// internal/services/user_service.go - 用户服务
package services

import (
    "context"
    "errors"
    "time"
    
    "github.com/your-org/autocodeweb/internal/models"
    "github.com/your-org/autocodeweb/internal/repositories"
    "github.com/your-org/autocodeweb/pkg/auth"
    "github.com/your-org/autocodeweb/pkg/cache"
)

type UserService struct {
    userRepo   repositories.UserRepository
    cache      cache.Cache
    authHelper auth.Helper
}

func NewUserService(userRepo repositories.UserRepository, cache cache.Cache, authHelper auth.Helper) *UserService {
    return &UserService{
        userRepo:   userRepo,
        cache:      cache,
        authHelper: authHelper,
    }
}

func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
    // 检查用户是否已存在
    existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err == nil && existingUser != nil {
        return nil, errors.New("user already exists")
    }
    
    // 创建用户
    user := &models.User{
        Email:     req.Email,
        Password:  s.authHelper.HashPassword(req.Password),
        Name:      req.Name,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // 清除缓存
    s.cache.Delete("users")
    
    return user, nil
}

func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
    // 查找用户
    user, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil || user == nil {
        return nil, errors.New("invalid credentials")
    }
    
    // 验证密码
    if !s.authHelper.CheckPassword(req.Password, user.Password) {
        return nil, errors.New("invalid credentials")
    }
    
    // 生成JWT token
    token, err := s.authHelper.GenerateToken(user.ID, user.Email)
    if err != nil {
        return nil, err
    }
    
    // 更新最后登录时间
    user.LastLoginAt = time.Now()
    s.userRepo.Update(ctx, user)
    
    return &models.LoginResponse{
        User:  user,
        Token: token,
    }, nil
}
```

### 4.2 项目服务架构
```go
// internal/services/project_service.go - 项目服务
package services

import (
    "context"
    "time"
    
    "github.com/your-org/autocodeweb/internal/models"
    "github.com/your-org/autocodeweb/internal/repositories"
    "github.com/your-org/autocodeweb/pkg/bmad"
)

type ProjectService struct {
    projectRepo repositories.ProjectRepository
    bmadEngine  bmad.Engine
}

func NewProjectService(projectRepo repositories.ProjectRepository, bmadEngine bmad.Engine) *ProjectService {
    return &ProjectService{
        projectRepo: projectRepo,
        bmadEngine:  bmadEngine,
    }
}

func (s *ProjectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.Project, error) {
    // 创建项目
    project := &models.Project{
        Name:        req.Name,
        Description: req.Description,
        UserID:      userID,
        Status:      models.ProjectStatusDraft,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := s.projectRepo.Create(ctx, project); err != nil {
        return nil, err
    }
    
    // 启动BMad-Method流程
    go func() {
        if err := s.bmadEngine.StartProject(context.Background(), project.ID, req.Requirements); err != nil {
            // 记录错误日志
            log.Printf("Failed to start BMad-Method flow for project %s: %v", project.ID, err)
        }
    }()
    
    return project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, projectID, userID string) (*models.Project, error) {
    project, err := s.projectRepo.FindByID(ctx, projectID)
    if err != nil {
        return nil, err
    }
    
    // 检查权限
    if project.UserID != userID {
        return nil, errors.New("access denied")
    }
    
    return project, nil
}

```

## 5. 中间件架构设计

### 5.1 认证中间件
```go
// internal/api/middleware/auth.go - 认证中间件
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/your-org/autocodeweb/pkg/auth"
)

func AuthMiddleware(authHelper auth.Helper) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取Authorization头
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // 解析Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }
        
        token := parts[1]
        
        // 验证token
        claims, err := authHelper.ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // 将用户信息存储到上下文中
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        
        c.Next()
    }
}

func OptionalAuthMiddleware(authHelper auth.Helper) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Next()
            return
        }
        
        parts := strings.Split(authHeader, " ")
        if len(parts) == 2 && parts[0] == "Bearer" {
            if claims, err := authHelper.ValidateToken(parts[1]); err == nil {
                c.Set("user_id", claims.UserID)
                c.Set("user_email", claims.Email)
            }
        }
        
        c.Next()
    }
}
```

### 5.2 限流中间件
```go
// internal/api/middleware/rate_limit.go - 限流中间件
package middleware

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/your-org/autocodeweb/pkg/cache"
)

type RateLimiter struct {
    cache cache.Cache
    limit int
    window time.Duration
}

func NewRateLimiter(cache cache.Cache, limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        cache:  cache,
        limit:  limit,
        window: window,
    }
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取客户端IP
        clientIP := c.ClientIP()
        key := "rate_limit:" + clientIP
        
        // 获取当前请求次数
        count, err := rl.cache.GetInt(key)
        if err != nil {
            count = 0
        }
        
        // 检查是否超过限制
        if count >= rl.limit {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": rl.window.Seconds(),
            })
            c.Abort()
            return
        }
        
        // 增加计数
        rl.cache.SetInt(key, count+1, rl.window)
        
        c.Next()
    }
}
```

## 6. 缓存架构设计

### 6.1 Redis缓存管理
```go
// pkg/cache/redis.go - Redis缓存管理
package cache

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
    return &RedisCache{client: client}
}

func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return c.client.Set(context.Background(), key, data, expiration).Err()
}

func (c *RedisCache) Get(key string, dest interface{}) error {
    data, err := c.client.Get(context.Background(), key).Bytes()
    if err != nil {
        return err
    }
    
    return json.Unmarshal(data, dest)
}

func (c *RedisCache) Delete(key string) error {
    return c.client.Del(context.Background(), key).Err()
}

func (c *RedisCache) Exists(key string) bool {
    result, err := c.client.Exists(context.Background(), key).Result()
    return err == nil && result > 0
}

func (c *RedisCache) SetInt(key string, value int, expiration time.Duration) error {
    return c.client.Set(context.Background(), key, value, expiration).Err()
}

func (c *RedisCache) GetInt(key string) (int, error) {
    result, err := c.client.Get(context.Background(), key).Int()
    if err != nil {
        return 0, err
    }
    return result, nil
}
```

## 7. 后台任务架构

### 7.1 任务队列管理
```go
// internal/worker/task_queue.go - 任务队列管理
package worker

import (
    "context"
    "encoding/json"
    "log"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/your-org/autocodeweb/internal/models"
)

type TaskQueue struct {
    redis *redis.Client
}

func NewTaskQueue(redis *redis.Client) *TaskQueue {
    return &TaskQueue{redis: redis}
}

func (tq *TaskQueue) EnqueueTask(ctx context.Context, task *models.Task) error {
    taskData, err := json.Marshal(task)
    if err != nil {
        return err
    }
    
    // 添加到Redis Stream
    return tq.redis.XAdd(ctx, &redis.XAddArgs{
        Stream: "task_queue",
        Values: map[string]interface{}{
            "task_data": string(taskData),
            "timestamp": time.Now().Unix(),
        },
    }).Err()
}

func (tq *TaskQueue) DequeueTask(ctx context.Context) (*models.Task, error) {
    // 从Redis Stream读取任务
    result, err := tq.redis.XRead(ctx, &redis.XReadArgs{
        Streams: []string{"task_queue", "0"},
        Count:   1,
        Block:   5 * time.Second,
    }).Result()
    
    if err != nil {
        return nil, err
    }
    
    if len(result) == 0 || len(result[0].Messages) == 0 {
        return nil, nil
    }
    
    message := result[0].Messages[0]
    taskData := message.Values["task_data"].(string)
    
    var task models.Task
    if err := json.Unmarshal([]byte(taskData), &task); err != nil {
        return nil, err
    }
    
    // 删除已处理的任务
    tq.redis.XDel(ctx, "task_queue", message.ID)
    
    return &task, nil
}
```

## 8. 错误处理架构

### 8.1 统一错误处理
```go
// pkg/errors/errors.go - 错误处理
package errors

import (
    "fmt"
    "net/http"
)

type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func NewBadRequest(message string) *AppError {
    return &AppError{
        Code:    http.StatusBadRequest,
        Message: message,
    }
}

func NewUnauthorized(message string) *AppError {
    return &AppError{
        Code:    http.StatusUnauthorized,
        Message: message,
    }
}

func NewNotFound(message string) *AppError {
    return &AppError{
        Code:    http.StatusNotFound,
        Message: message,
    }
}

func NewInternalServerError(message string) *AppError {
    return &AppError{
        Code:    http.StatusInternalServerError,
        Message: message,
    }
}
```

---

*本文档为 AutoCodeWeb 项目的后端架构设计，由架构师 Winston 创建*
