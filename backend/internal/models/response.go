package models

import (
	"time"
)

// PaginationResponse 分页响应
type PaginationResponse struct {
	Code        int         `json:"code" example:"0"`
	Message     string      `json:"message" example:"success"`
	Total       int         `json:"total" example:"100"`
	Page        int         `json:"page" example:"1"`
	PageSize    int         `json:"page_size" example:"10"`
	TotalPages  int         `json:"total_pages" example:"10"`
	Data        interface{} `json:"data"`
	HasNext     bool        `json:"has_next" example:"true"`
	HasPrevious bool        `json:"has_previous" example:"false"`
	Timestamp   string      `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// UserInfo 用户信息（用于响应）
type UserInfo struct {
	ID        string    `json:"id" example:"varchar(50)"`
	Email     string    `json:"email" example:"user@example.com"`
	Username  string    `json:"username" example:"username"`
	Role      string    `json:"role" example:"user"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectInfo 项目信息（用于响应）
type ProjectInfo struct {
	GUID         string    `json:"guid" example:"e080335a93d0456ba9b65ab407710e55"`
	Name         string    `json:"name" example:"项目名称"`
	Description  string    `json:"description" example:"项目描述"`
	Status       string    `json:"status" example:"in_progress"`
	Requirements string    `json:"requirements" example:"项目需求"`
	ProjectPath  string    `json:"project_path" example:"/path/to/project"`
	BackendPort  int       `json:"backend_port" example:"8080"`
	FrontendPort int       `json:"frontend_port" example:"3000"`
	UserID       string    `json:"user_id" example:"USER_00000000001"`
	User         UserInfo  `json:"user,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// convertToProjectInfo 将Project模型转换为ProjectInfo响应格式
func ConvertToProjectInfo(project *Project) *ProjectInfo {
	projectInfo := &ProjectInfo{
		GUID:         project.GUID,
		Name:         project.Name,
		Description:  project.Description,
		Status:       project.Status,
		Requirements: project.Requirements,
		ProjectPath:  project.ProjectPath,
		BackendPort:  project.BackendPort,
		FrontendPort: project.FrontendPort,
		UserID:       project.UserID,
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}

	// 转换用户信息
	if project.User.ID != "" {
		projectInfo.User = UserInfo{
			ID:        project.User.ID,
			Email:     project.User.Email,
			Username:  project.User.Username,
			Role:      project.User.Role,
			Status:    project.User.Status,
			CreatedAt: project.User.CreatedAt,
		}
	}

	return projectInfo
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in" example:"3600"`
}

// UserSettingsResponse 用户设置响应
type UserSettingsResponse struct {
	DefaultCliTool       string `json:"default_cli_tool" example:"claude-code"`
	DefaultAiModel       string `json:"default_ai_model" example:"glm-4.6"`
	DefaultModelProvider string `json:"default_model_provider" example:"zhipu"`
	DefaultModelApiUrl   string `json:"default_model_api_url" example:"https://open.bigmodel.cn/api/anthropic"`
	DefaultApiToken      string `json:"default_api_token,omitempty" example:"sk-***"` // 敏感信息，前端可能需要脱敏显示
}
