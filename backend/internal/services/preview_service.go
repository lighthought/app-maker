package services

import (
	"context"
	"fmt"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"

	"shared-models/logger"
)

type PreviewService interface {
	GeneratePreviewToken(ctx context.Context, projectID string, expiresInDays int) (*models.PreviewToken, error)
	GetPreviewByToken(ctx context.Context, token string) (*models.PreviewToken, error)
	GetProjectTokens(ctx context.Context, projectID string) ([]*models.PreviewToken, error)
	DeleteToken(ctx context.Context, id string) error
	CleanupExpiredTokens(ctx context.Context) error
}

type previewService struct {
	tokenRepo repositories.PreviewTokenRepository
}

func NewPreviewService(tokenRepo repositories.PreviewTokenRepository) PreviewService {
	return &previewService{
		tokenRepo: tokenRepo,
	}
}

func (s *previewService) GeneratePreviewToken(ctx context.Context, projectID string, expiresInDays int) (*models.PreviewToken, error) {
	if expiresInDays <= 0 {
		expiresInDays = 7 // 默认7天
	}

	token := &models.PreviewToken{
		ProjectID: projectID,
		ExpiresAt: time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour),
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		logger.Error("生成预览令牌失败",
			logger.String("projectID", projectID),
			logger.String("error", err.Error()),
		)
		return nil, fmt.Errorf("生成预览令牌失败: %w", err)
	}

	logger.Info("生成预览令牌成功",
		logger.String("projectID", projectID),
		logger.String("tokenID", token.ID),
	)

	return token, nil
}

func (s *previewService) GetPreviewByToken(ctx context.Context, token string) (*models.PreviewToken, error) {
	previewToken, err := s.tokenRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("获取预览令牌失败: %w", err)
	}

	if previewToken.IsExpired() {
		return nil, fmt.Errorf("预览令牌已过期")
	}

	return previewToken, nil
}

func (s *previewService) GetProjectTokens(ctx context.Context, projectID string) ([]*models.PreviewToken, error) {
	return s.tokenRepo.GetByProjectID(ctx, projectID)
}

func (s *previewService) DeleteToken(ctx context.Context, id string) error {
	return s.tokenRepo.Delete(ctx, id)
}

func (s *previewService) CleanupExpiredTokens(ctx context.Context) error {
	return s.tokenRepo.DeleteExpired(ctx)
}
