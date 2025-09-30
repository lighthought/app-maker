package services

import (
	"context"
	"errors"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"shared-models/auth"
	"shared-models/common"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务接口
type UserService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, userID, token string) error
	GetUserProfile(ctx context.Context, userID string) (*models.UserInfo, error)
	UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) error
	ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error
	GetUserList(ctx context.Context, page, pageSize int) (*models.PaginationResponse, error)
	DeleteUser(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo       repositories.UserRepository
	authJWTService *auth.JWTService
	jwtExpireHours int
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repositories.UserRepository, jwtService *auth.JWTService, jwtExpireHours int) UserService {
	return &userService{
		userRepo:       userRepo,
		authJWTService: jwtService,
		jwtExpireHours: jwtExpireHours,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(common.MESSAGE_EMAIL_ALREADY_EXISTS)
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(common.MESSAGE_USERNAME_ALREADY_EXISTS)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     common.UserRoleUser,
		Status:   common.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User: models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtExpireHours * 3600),
	}, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// 根据邮箱获取用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New(common.MESSAGE_USER_OR_PASSWORD_ERROR)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New(common.MESSAGE_USER_OR_PASSWORD_ERROR)
	}

	// 检查用户状态
	if user.Status != common.UserStatusActive {
		return nil, errors.New(common.MESSAGE_USER_DISABLED)
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User: models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtExpireHours * 3600),
	}, nil
}

// Logout 用户登出
func (s *userService) Logout(ctx context.Context, userID, token string) error {
	// 对于纯JWT实现，登出只需要客户端清除token
	// 这里可以添加token黑名单逻辑，但为了简化，暂时不实现
	return nil
}

// GetUserProfile 获取用户档案
func (s *userService) GetUserProfile(ctx context.Context, userID string) (*models.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New(common.MESSAGE_USER_NOT_FOUND)
	}

	return &models.UserInfo{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}, nil
}

// UpdateUserProfile 更新用户档案
func (s *userService) UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New(common.MESSAGE_USER_NOT_FOUND)
	}

	// 更新字段
	if req.Username != "" {
		// 检查用户名是否已被其他用户使用
		if req.Username != user.Username {
			exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
			if err != nil {
				return err
			}
			if exists {
				return errors.New(common.MESSAGE_USERNAME_ALREADY_EXISTS)
			}
			user.Username = req.Username
		}
	}

	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		if req.Email != user.Email {
			exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
			if err != nil {
				return err
			}
			if exists {
				return errors.New(common.MESSAGE_EMAIL_ALREADY_EXISTS)
			}
			user.Email = req.Email
		}
	}

	return s.userRepo.Update(ctx, user)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New(common.MESSAGE_USER_NOT_FOUND)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New(common.MESSAGE_OLD_PASSWORD_ERROR)
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// GetUserList 获取用户列表
func (s *userService) GetUserList(ctx context.Context, page, pageSize int) (*models.PaginationResponse, error) {
	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为UserInfo
	userInfos := make([]models.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		}
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &models.PaginationResponse{
		Total:       int(total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Data:        userInfos,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	// 删除用户
	return s.userRepo.Delete(ctx, userID)
}

// RefreshToken 刷新令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// 创建 JWT 服务来验证刷新令牌
	// 验证刷新令牌
	userID, err := s.authJWTService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New(common.MESSAGE_INVALID_REFRESH_TOKEN)
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New(common.MESSAGE_USER_NOT_FOUND)
	}

	// 检查用户状态
	if user.Status != common.UserStatusActive {
		return nil, errors.New(common.MESSAGE_USER_DISABLED)
	}

	// 生成新的令牌
	accessToken, newRefreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User: models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtExpireHours * 3600),
	}, nil
}

// generateTokens 生成JWT令牌
func (s *userService) generateTokens(userID string) (string, string, error) {
	// 获取用户信息用于生成 JWT
	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		return "", "", err
	}

	// 使用 JWT 服务生成令牌
	return s.authJWTService.GenerateTokens(user.ID, user.Email, user.Username)
}
