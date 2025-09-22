package models

import (
	"time"

	"gorm.io/gorm"
)

// ConversationMessage 对话消息模型
type ConversationMessage struct {
	ID              string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('MSG', 'public.project_msgs_id_num_seq')"`
	ProjectGuid     string         `json:"project_guid" gorm:"type:varchar(50);"`
	Type            string         `json:"type" gorm:"size:20;not null"` // user, agent, system
	AgentRole       string         `json:"agent_role" gorm:"size:20"`    // dev, pm, arch, ux, qa, ops
	AgentName       string         `json:"agent_name" gorm:"size:100"`   // Agent名称
	Content         string         `json:"content" gorm:"type:text"`
	IsMarkdown      bool           `json:"is_markdown" gorm:"default:false"`
	MarkdownContent string         `json:"markdown_content" gorm:"type:text"`
	IsExpanded      bool           `json:"is_expanded" gorm:"default:false"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ConversationMessage) TableName() string {
	return "project_msgs"
}

// BeforeCreate 创建前的钩子
func (cm *ConversationMessage) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('MSG', 'project_msgs_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		cm.ID = result
	}
	return nil
}
