package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务模型
type Task struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('TASK', 'public.tasks_id_num_seq')"`
	ProjectID   string         `json:"project_id" gorm:"type:varchar(50);not null"`
	Type        string         `json:"type" gorm:"size:50;not null"`                     // project_development, build, deploy, etc.
	Status      string         `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, completed, failed, cancelled
	Priority    int            `json:"priority" gorm:"default:0"`                        // 0-9, 0为最高优先级
	Description string         `json:"description" gorm:"type:text"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Project Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Logs    []TaskLog `json:"logs,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskLog 任务日志模型
type TaskLog struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('LOGS', 'public.task_logs_id_num_seq')"`
	TaskID    string    `json:"task_id" gorm:"type:varchar(50);not null"`
	Level     string    `json:"level" gorm:"size:10;not null"` // info, success, warning, error
	Message   string    `json:"message" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联关系
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

func (TaskLog) TableName() string {
	return "task_logs"
}

// BeforeCreate 创建前的钩子
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.Status == "" {
		t.Status = "pending"
	}
	if t.Priority == 0 {
		t.Priority = 5
	}
	if t.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('TASK', 'tasks_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		t.ID = result
	}
	return nil
}
