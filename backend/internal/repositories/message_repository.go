package repositories

import (
	"autocodeweb-backend/internal/models"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// MessageRepository 对话消息仓库接口
type MessageRepository interface {
	// Create 创建对话消息
	Create(ctx context.Context, message *models.ConversationMessage) error

	// GetByProjectGuid 获取项目的对话消息列表
	GetByProjectGuid(ctx context.Context, projectGuid string, limit, offset int) ([]*models.ConversationMessage, error)

	// GetByID 根据ID获取对话消息
	GetByID(ctx context.Context, id string) (*models.ConversationMessage, error)

	// Update 更新对话消息
	Update(ctx context.Context, message *models.ConversationMessage) error

	// Delete 删除对话消息
	Delete(ctx context.Context, id string) error

	// CountByProjectGuid 统计项目的对话消息数量
	CountByProjectGuid(ctx context.Context, projectGuid string) (int, error)
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

func (r *messageRepository) GetByProjectGuid(ctx context.Context, projectGuid string, limit, offset int) ([]*models.ConversationMessage, error) {
	var messages []*models.ConversationMessage
	err := r.db.WithContext(ctx).
		Where("project_guid = ?", projectGuid).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	// 添加调试日志
	fmt.Printf("DEBUG: Query project_msgs for projectGuid=%s, limit=%d, offset=%d, found %d messages\n",
		projectGuid, limit, offset, len(messages))

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

func (r *messageRepository) CountByProjectGuid(ctx context.Context, projectGuid string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ConversationMessage{}).
		Where("project_guid = ?", projectGuid).
		Count(&count).Error
	return int(count), err
}
