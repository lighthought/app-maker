package models

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"not null"`
	Description  string         `json:"description"`
	UserID       string         `json:"user_id" gorm:"not null;type:uuid"`
	Status       string         `json:"status" gorm:"default:'draft';check:status IN ('draft', 'in_progress', 'completed', 'failed')"`
	Requirements string         `json:"requirements" gorm:"type:text"`
	ProjectPath  string         `json:"project_path" gorm:"uniqueIndex"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	User  User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ProjectID"`
	Tags  []Tag  `json:"tags,omitempty" gorm:"many2many:project_tags;"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// BeforeCreate 创建前的钩子
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = "draft"
	}
	return nil
}
