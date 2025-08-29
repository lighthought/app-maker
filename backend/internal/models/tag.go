package models

import (
	"time"

	"gorm.io/gorm"
)

// Tag 标签模型
type Tag struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"`
	Color     string         `json:"color" gorm:"default:'#666666'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_tags;"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// ProjectTag 项目标签关联表
type ProjectTag struct {
	ProjectID string    `json:"project_id" gorm:"primaryKey;type:uuid"`
	TagID     string    `json:"tag_id" gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProjectTag) TableName() string {
	return "project_tags"
}
