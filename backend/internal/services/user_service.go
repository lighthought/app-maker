package services

import (
	"context"
	"errors"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"

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
	userRepo        repositories.UserRepository
	userSessionRepo repositories.UserSessionRepository
	jwtSecret       string
	jwtExpireHours  int
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repositories.UserRepository, userSessionRepo repositories.UserSessionRepository, jwtSecret string, jwtExpireHours int) UserService {
	return &userService{
		userRepo:        userRepo,
		userSessionRepo: userSessionRepo,
		jwtSecret:       jwtSecret,
		jwtExpireHours:  jwtExpireHours,
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
		return nil, errors.New("邮箱已存在")
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
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
		Role:     "user",
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	// 创建用户会话
	session := &models.UserSession{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(s.jwtExpireHours*2) * time.Hour), // 刷新令牌有效期更长
	}

	if err := s.userSessionRepo.Create(ctx, session); err != nil {
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
		return nil, errors.New("用户不存在或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户账户已被禁用")
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	// 检查是否已存在会话，如果存在则更新，否则创建新的
	existingSessions, err := s.userSessionRepo.GetByUserID(ctx, user.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}

	if len(existingSessions) > 0 {
		// 更新现有会话
		session := &existingSessions[0]
		session.Token = refreshToken
		session.ExpiresAt = time.Now().Add(time.Duration(s.jwtExpireHours*2) * time.Hour)
		if err := s.userSessionRepo.Update(ctx, session); err != nil {
			return nil, err
		}
	} else {
		// 创建新的用户会话
		session := &models.UserSession{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().Add(time.Duration(s.jwtExpireHours*2) * time.Hour),
		}

		if err := s.userSessionRepo.Create(ctx, session); err != nil {
			return nil, err
		}
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
	// 删除当前会话
	return s.userSessionRepo.Delete(ctx, token)
}

// GetUserProfile 获取用户档案
func (s *userService) GetUserProfile(ctx context.Context, userID string) (*models.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
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
		return errors.New("用户不存在")
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
				return errors.New("用户名已被使用")
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
				return errors.New("邮箱已被使用")
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
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("旧密码错误")
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
	// 删除用户的所有会话
	if err := s.userSessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	// 删除用户
	return s.userRepo.Delete(ctx, userID)
}

// RefreshToken 刷新令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// 验证刷新令牌
	session, err := s.userSessionRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("无效的刷新令牌")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户账户已被禁用")
	}

	// 生成新的令牌
	accessToken, newRefreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	// 更新会话令牌
	session.Token = newRefreshToken
	session.ExpiresAt = time.Now().Add(time.Duration(s.jwtExpireHours*2) * time.Hour)
	if err := s.userSessionRepo.Update(ctx, session); err != nil {
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
	// 这里应该实现JWT令牌生成逻辑
	// 为了简化，暂时返回模拟的令牌
	accessToken := "access_token_" + userID
	refreshToken := "refresh_token_" + userID
	return accessToken, refreshToken, nil
}
