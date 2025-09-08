package models

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"min=1" example:"1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100" example:"10"`
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

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Requirements string `json:"requirements" binding:"required" example:"项目需求描述"`
}

// ProjectListRequest 项目列表请求
type ProjectListRequest struct {
	PaginationRequest
	Status string   `json:"status" form:"status" binding:"omitempty,oneof=draft in_progress completed failed" example:"in_progress"`
	TagIDs []string `json:"tag_ids" form:"tag_ids" example:"['tag1', 'tag2']"`
	UserID string   `json:"user_id" form:"user_id" example:"USER_00000000001"`
	Search string   `json:"search" form:"search" example:"项目名称关键词"`
}

// JenkinsBuildRequest Jenkins 构建请求
type JenkinsBuildRequest struct {
	UserID      string `json:"user_id"`
	ProjectID   string `json:"project_id"`
	ProjectPath string `json:"project_path"`
	BuildType   string `json:"build_type"` // dev 或 prod
}
