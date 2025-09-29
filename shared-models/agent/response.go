package agent

// AgentHealthResp Agent 健康检查响应
type AgentHealthResp struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// 项目环境准备响应
type SetupProjEnvResp struct {
	BmadMethodStatus string `json:"bmad_method_status" example:"success"`
	FrontendStatus   string `json:"frontend_status" example:"success"`
	BackendStatus    string `json:"backend_status" example:"success"`
}
