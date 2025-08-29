package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务模型
type Task struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	ProjectID   string         `json:"project_id" gorm:"not null;type:uuid"`
	Type        string         `json:"type" gorm:"not null;check:type IN ('prd_generation', 'architecture_design', 'ux_design', 'epic_story', 'coding', 'testing', 'deployment')"`
	Status      string         `json:"status" gorm:"default:'pending';check:status IN ('pending', 'running', 'completed', 'failed', 'cancelled')"`
	Priority    string         `json:"priority" gorm:"default:'normal';check:priority IN ('low', 'normal', 'high', 'urgent')"`
	Progress    int            `json:"progress" gorm:"default:0;check:progress >= 0 AND progress <= 100"`
	Result      string         `json:"result" gorm:"type:text"`
	Error       string         `json:"error" gorm:"type:text"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// BeforeCreate 创建前的钩子
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.Status == "" {
		t.Status = "pending"
	}
	if t.Priority == "" {
		t.Priority = "normal"
	}
	if t.Progress == 0 {
		t.Progress = 0
	}
	return nil
}
