package models

import (
	"time"

	"gorm.io/gorm"
)

// ConversationMessage 对话消息模型
type ConversationMessage struct {
	ID              string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('MSG', 'public.conversation_messages_id_num_seq')"`
	ProjectID       string         `json:"project_id" gorm:"type:varchar(50);not null"`
	Project         Project        `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
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

// DevStage 开发阶段模型
type DevStage struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('STAGE', 'public.dev_stages_id_num_seq')"`
	ProjectID   string         `json:"project_id" gorm:"type:varchar(50);not null"`
	Project     Project        `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Status      string         `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, completed, failed
	Progress    int            `json:"progress" gorm:"default:0"`                        // 0-100
	Description string         `json:"description" gorm:"type:text"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ConversationMessage) TableName() string {
	return "conversation_messages"
}

func (DevStage) TableName() string {
	return "dev_stages"
}

// BeforeCreate 创建前的钩子
func (cm *ConversationMessage) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('MSG', 'conversation_messages_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		cm.ID = result
	}
	return nil
}

func (ds *DevStage) BeforeCreate(tx *gorm.DB) error {
	if ds.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('STAGE', 'dev_stages_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		ds.ID = result
	}
	return nil
}
