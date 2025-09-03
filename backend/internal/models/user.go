package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('USER', 'public.users_id_num_seq')"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // 不在JSON中显示密码
	Role      string         `json:"role" gorm:"default:'user';check:role IN ('admin', 'user')"`
	Status    string         `json:"status" gorm:"default:'active';check:status IN ('active', 'inactive', 'suspended')"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		// 手动生成ID
		var result string
		err := tx.Raw("SELECT generate_table_id('USER', 'users_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		u.ID = result
	}
	return nil
}
