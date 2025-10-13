package models

import (
	"time"

	"shared-models/utils"

	"gorm.io/gorm"
)

// PreviewToken 预览令牌模型
type PreviewToken struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(50)"`
	ProjectID string         `json:"project_id" gorm:"type:varchar(50);not null;index"`
	Token     string         `json:"token" gorm:"size:100;uniqueIndex;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Project   *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// TableName 指定表名
func (PreviewToken) TableName() string {
	return "preview_tokens"
}

// BeforeCreate 创建前的钩子
func (p *PreviewToken) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = utils.GenerateUUID()
	}
	if p.Token == "" {
		p.Token = utils.GenerateUUID()
	}
	// 默认7天过期
	if p.ExpiresAt.IsZero() {
		p.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	return nil
}

// IsExpired 检查令牌是否已过期
func (p *PreviewToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}
