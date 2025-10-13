package repositories

import (
	"context"
	"time"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

type PreviewTokenRepository interface {
	Create(ctx context.Context, token *models.PreviewToken) error
	GetByToken(ctx context.Context, tokenStr string) (*models.PreviewToken, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*models.PreviewToken, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

type previewTokenRepository struct {
	db *gorm.DB
}

func NewPreviewTokenRepository(db *gorm.DB) PreviewTokenRepository {
	return &previewTokenRepository{db: db}
}

func (r *previewTokenRepository) Create(ctx context.Context, token *models.PreviewToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *previewTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*models.PreviewToken, error) {
	var token models.PreviewToken
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("token = ? AND expires_at > ?", tokenStr, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *previewTokenRepository) GetByProjectID(ctx context.Context, projectID string) ([]*models.PreviewToken, error) {
	var tokens []*models.PreviewToken
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND expires_at > ?", projectID, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

func (r *previewTokenRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.PreviewToken{}, "id = ?", id).Error
}

func (r *previewTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&models.PreviewToken{}).Error
}
