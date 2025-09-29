package agent

// 项目环境准备响应
type SetupProjEnvRes struct {
	BmadMethodStatus string `json:"bmad_method_status" example:"success"`
	FrontendStatus   string `json:"frontend_status" example:"success"`
	BackendStatus    string `json:"backend_status" example:"success"`
}
