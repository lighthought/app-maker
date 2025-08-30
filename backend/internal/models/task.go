package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"     // 待执行
	TaskStatusRunning    TaskStatus = "running"     // 执行中
	TaskStatusCompleted  TaskStatus = "completed"   // 已完成
	TaskStatusFailed     TaskStatus = "failed"      // 执行失败
	TaskStatusCancelled  TaskStatus = "cancelled"   // 已取消
	TaskStatusRetrying   TaskStatus = "retrying"    // 重试中
	TaskStatusRolledBack TaskStatus = "rolled_back" // 已回滚
)

// TaskPriority 任务优先级枚举
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 1
	TaskPriorityNormal TaskPriority = 2
	TaskPriorityHigh   TaskPriority = 3
	TaskPriorityUrgent TaskPriority = 4
)

// Task 任务模型
type Task struct {
	ID          string       `json:"id" gorm:"primaryKey;type:uuid"`
	ProjectID   string       `json:"project_id" gorm:"type:uuid;not null"`
	UserID      string       `json:"user_id" gorm:"type:uuid;not null"`
	Name        string       `json:"name" gorm:"not null"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Priority    TaskPriority `json:"priority" gorm:"default:2"`

	// 依赖关系
	Dependencies []string `json:"dependencies" gorm:"type:text[]"` // 依赖的任务ID列表

	// 执行相关
	MaxRetries int `json:"max_retries" gorm:"default:3"`
	RetryCount int `json:"retry_count" gorm:"default:0"`
	RetryDelay int `json:"retry_delay" gorm:"default:60"` // 重试延迟（秒）

	// 时间相关
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Deadline    *time.Time `json:"deadline"`

	// 执行结果
	Result       string `json:"result"` // 执行结果数据
	ErrorMessage string `json:"error_message"`

	// 元数据
	Metadata string   `json:"metadata"` // JSON格式的元数据
	Tags     []string `json:"tags" gorm:"type:text[]"`

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	Project Project   `json:"project" gorm:"foreignKey:ProjectID"`
	User    User      `json:"user" gorm:"foreignKey:UserID"`
	Logs    []TaskLog `json:"logs" gorm:"foreignKey:TaskID"`
}

// TaskLog 任务执行日志
type TaskLog struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	TaskID    string    `json:"task_id" gorm:"type:uuid;not null"`
	Level     string    `json:"level" gorm:"type:varchar(10);not null"` // info, warn, error
	Message   string    `json:"message" gorm:"not null"`
	Data      string    `json:"data"` // JSON格式的额外数据
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	Task Task `json:"task" gorm:"foreignKey:TaskID"`
}

// TaskDependency 任务依赖关系
type TaskDependency struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid"`
	TaskID       string    `json:"task_id" gorm:"type:uuid;not null"`
	DependencyID string    `json:"dependency_id" gorm:"type:uuid;not null"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联关系
	Task       Task `json:"task" gorm:"foreignKey:TaskID"`
	Dependency Task `json:"dependency" gorm:"foreignKey:DependencyID"`
}

// BeforeCreate 创建前钩子
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 创建前钩子
func (tl *TaskLog) BeforeCreate(tx *gorm.DB) error {
	if tl.ID == "" {
		tl.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 创建前钩子
func (td *TaskDependency) BeforeCreate(tx *gorm.DB) error {
	if td.ID == "" {
		td.ID = uuid.New().String()
	}
	return nil
}
