package models

import (
	"time"

	"github.com/lighthought/app-maker/shared-models/common"

	"gorm.io/gorm"
)

// ConversationMessage 对话消息模型
type ConversationMessage struct {
	ID                  string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('MSG', 'public.project_msgs_id_num_seq')"`
	ProjectGuid         string         `json:"project_guid" gorm:"type:varchar(50);"`
	Type                string         `json:"type" gorm:"size:20;not null"` // user, agent, system
	AgentRole           string         `json:"agent_role" gorm:"size:20"`    // user, dev, pm, arch, ux, qa, ops
	AgentName           string         `json:"agent_name" gorm:"size:100"`   // Agent名称
	Content             string         `json:"content" gorm:"type:text"`
	IsMarkdown          bool           `json:"is_markdown" gorm:"default:false"`
	MarkdownContent     string         `json:"markdown_content" gorm:"type:text"`
	IsExpanded          bool           `json:"is_expanded" gorm:"default:false"`
	HasQuestion         bool           `json:"has_question" gorm:"default:false"`          // 是否包含问题
	WaitingUserResponse bool           `json:"waiting_user_response" gorm:"default:false"` // 是否等待用户回复
	CreatedAt           time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ConversationMessage) TableName() string {
	return "project_msgs"
}

// Copy 复制对话消息
func (c *ConversationMessage) Copy(other *ConversationMessage) {
	c.ID = other.ID
	c.ProjectGuid = other.ProjectGuid
	c.Type = other.Type
	c.AgentRole = other.AgentRole
	c.AgentName = other.AgentName
	c.Content = other.Content
	c.IsMarkdown = other.IsMarkdown
	c.MarkdownContent = other.MarkdownContent
	c.IsExpanded = other.IsExpanded
	c.HasQuestion = other.HasQuestion
	c.WaitingUserResponse = other.WaitingUserResponse
	c.CreatedAt = other.CreatedAt
	c.UpdatedAt = other.UpdatedAt
	c.DeletedAt = other.DeletedAt
}

// NewUserMessage 创建用户消息
func NewUserMessage(project *Project) *ConversationMessage {
	return &ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeUser,
		AgentRole:       common.AgentTypeUser,
		AgentName:       common.AgentTypeUser,
		Content:         project.Requirements,
		IsMarkdown:      false,
		MarkdownContent: project.Requirements,
		IsExpanded:      true,
	}
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
