package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // 不在JSON中显示密码
	Role      string         `json:"role" gorm:"default:'user';check:role IN ('admin', 'user')"`
	Status    string         `json:"status" gorm:"default:'active';check:status IN ('active', 'inactive', 'suspended')"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	UserSessions []UserSession `json:"user_sessions,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// UserSession 用户会话模型
type UserSession struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null"`
	Token     string         `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (UserSession) TableName() string {
	return "user_sessions"
}

// BeforeCreate 创建前的钩子
func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == "" {
		us.ID = uuid.New().String()
	}
	return nil
}
