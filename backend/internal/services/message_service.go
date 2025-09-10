package services

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// MessageService 对话消息服务接口
type MessageService interface {
	// GetProjectConversations 获取项目对话历史
	GetProjectConversations(ctx context.Context, projectID string, page, pageSize int) ([]*models.ConversationMessage, int, error)

	// AddConversationMessage 添加对话消息
	AddConversationMessage(ctx context.Context, projectID string, message *models.ConversationMessage) (*models.ConversationMessage, error)

	// GetConversationMessage 获取对话消息
	GetConversationMessage(ctx context.Context, messageID string) (*models.ConversationMessage, error)

	// UpdateConversationMessage 更新对话消息
	UpdateConversationMessage(ctx context.Context, message *models.ConversationMessage) error
}

// messageService 对话消息服务实现
type messageService struct {
	repo repositories.MessageRepository
}

// NewMessageService 创建对话消息服务
func NewMessageService(db *gorm.DB) MessageService {
	repo := repositories.NewMessageRepository(db)
	return &messageService{repo: repo}
}

func (s *messageService) GetProjectConversations(ctx context.Context, projectID string, page, pageSize int) ([]*models.ConversationMessage, int, error) {
	offset := (page - 1) * pageSize
	messages, err := s.repo.GetByProjectID(ctx, projectID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountByProjectID(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (s *messageService) AddConversationMessage(ctx context.Context, projectID string, message *models.ConversationMessage) (*models.ConversationMessage, error) {
	message.ProjectID = projectID
	if err := s.repo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("创建对话消息失败: %w", err)
	}
	return message, nil
}

func (s *messageService) GetConversationMessage(ctx context.Context, messageID string) (*models.ConversationMessage, error) {
	return s.repo.GetByID(ctx, messageID)
}

func (s *messageService) UpdateConversationMessage(ctx context.Context, message *models.ConversationMessage) error {
	return s.repo.Update(ctx, message)
}
