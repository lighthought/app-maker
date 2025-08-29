package models

import (
	"time"
)

// Response 通用响应结构
type Response struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int    `json:"code" example:"400"`
	Message   string `json:"message" example:"请求参数错误"`
	Timestamp string `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"min=1" example:"1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100" example:"10"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Total       int         `json:"total" example:"100"`
	Page        int         `json:"page" example:"1"`
	PageSize    int         `json:"page_size" example:"10"`
	TotalPages  int         `json:"total_pages" example:"10"`
	Data        interface{} `json:"data"`
	HasNext     bool        `json:"has_next" example:"true"`
	HasPrevious bool        `json:"has_previous" example:"false"`
}

// BaseModel 基础模型
type BaseModel struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
}

// UserInfo 用户信息（用于响应）
type UserInfo struct {
	ID        string    `json:"id" example:"uuid"`
	Email     string    `json:"email" example:"user@example.com"`
	Username  string    `json:"username" example:"username"`
	Role      string    `json:"role" example:"user"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Username string `json:"username" binding:"required,min=3,max=20" example:"username"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in" example:"3600"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// UpdateProfileRequest 更新用户档案请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=20" example:"newusername"`
	Email    string `json:"email" binding:"omitempty,email" example:"newemail@example.com"`
}
