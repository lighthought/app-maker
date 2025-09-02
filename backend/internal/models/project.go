package models

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID               string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name             string         `json:"name" gorm:"not null"`
	Description      string         `json:"description"`
	UserID           string         `json:"user_id" gorm:"not null;type:uuid"`
	Status           string         `json:"status" gorm:"default:'draft';check:status IN ('draft', 'in_progress', 'completed', 'failed')"`
	Requirements     string         `json:"requirements" gorm:"type:text"`
	ProjectPath      string         `json:"project_path" gorm:"uniqueIndex"`
	BackendPort      int            `json:"backend_port" gorm:"default:8080"`
	FrontendPort     int            `json:"frontend_port" gorm:"default:3000"`
	ApiBaseUrl       string         `json:"api_base_url" gorm:"default:'/api/v1'"`
	AppSecretKey     string         `json:"app_secret_key"`
	DatabasePassword string         `json:"database_password"`
	RedisPassword    string         `json:"redis_password"`
	JwtSecretKey     string         `json:"jwt_secret_key"`
	Subnetwork       string         `json:"subnetwork" gorm:"default:'172.20.0.0/16'"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

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
	if p.BackendPort == 0 {
		p.BackendPort = 8080
	}
	if p.FrontendPort == 0 {
		p.FrontendPort = 3000
	}
	return nil
}
