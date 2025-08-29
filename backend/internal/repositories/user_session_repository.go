package repositories

import (
	"context"
	"time"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

// UserSessionRepository 用户会话数据访问接口
type UserSessionRepository interface {
	Create(ctx context.Context, session *models.UserSession) error
	GetByToken(ctx context.Context, token string) (*models.UserSession, error)
	GetByUserID(ctx context.Context, userID string) ([]models.UserSession, error)
	Update(ctx context.Context, session *models.UserSession) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// userSessionRepository 用户会话数据访问实现
type userSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository 创建用户会话数据访问实例
func NewUserSessionRepository(db *gorm.DB) UserSessionRepository {
	return &userSessionRepository{db: db}
}

// Create 创建用户会话
func (r *userSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByToken 根据令牌获取会话
func (r *userSessionRepository) GetByToken(ctx context.Context, token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserID 根据用户ID获取会话列表
func (r *userSessionRepository) GetByUserID(ctx context.Context, userID string) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ? AND expires_at > ?", userID, time.Now()).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Update 更新用户会话
func (r *userSessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete 删除用户会话
func (r *userSessionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "id = ?", id).Error
}

// DeleteExpired 删除过期会话
func (r *userSessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "expires_at <= ?", time.Now()).Error
}

// DeleteByUserID 根据用户ID删除所有会话
func (r *userSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "user_id = ?", userID).Error
}

// CleanupExpiredSessions 清理过期会话
func (r *userSessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, "expires_at <= ?", time.Now()).Error
}
