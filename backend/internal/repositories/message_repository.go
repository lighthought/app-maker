package repositories

import (
	"autocodeweb-backend/internal/models"
	"context"

	"gorm.io/gorm"
)

// MessageRepository 对话消息仓库接口
type MessageRepository interface {
	// Create 创建对话消息
	Create(ctx context.Context, message *models.ConversationMessage) error

	// GetByProjectID 获取项目的对话消息列表
	GetByProjectID(ctx context.Context, projectID string, limit, offset int) ([]*models.ConversationMessage, error)

	// GetByID 根据ID获取对话消息
	GetByID(ctx context.Context, id string) (*models.ConversationMessage, error)

	// Update 更新对话消息
	Update(ctx context.Context, message *models.ConversationMessage) error

	// Delete 删除对话消息
	Delete(ctx context.Context, id string) error

	// CountByProjectID 统计项目的对话消息数量
	CountByProjectID(ctx context.Context, projectID string) (int, error)
}

// messageRepository 对话消息仓库实现
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建对话消息仓库
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.ConversationMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) GetByProjectID(ctx context.Context, projectID string, limit, offset int) ([]*models.ConversationMessage, error) {
	var messages []*models.ConversationMessage
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetByID(ctx context.Context, id string) (*models.ConversationMessage, error) {
	var message models.ConversationMessage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) Update(ctx context.Context, message *models.ConversationMessage) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ConversationMessage{}, "id = ?", id).Error
}

func (r *messageRepository) CountByProjectID(ctx context.Context, projectID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ConversationMessage{}).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return int(count), err
}
