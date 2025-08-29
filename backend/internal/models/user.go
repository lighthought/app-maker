package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email       string         `json:"email" gorm:"uniqueIndex;not null"`
	Password    string         `json:"-" gorm:"not null"` // 不在JSON中返回密码
	Name        string         `json:"name" gorm:"not null"`
	Role        string         `json:"role" gorm:"default:'user';check:role IN ('admin', 'user')"`
	Status      string         `json:"status" gorm:"default:'active';check:status IN ('active', 'inactive', 'suspended')"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Projects []Project `json:"projects,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	if u.Status == "" {
		u.Status = "active"
	}
	return nil
}
