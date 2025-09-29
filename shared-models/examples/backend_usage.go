package examples

import (
	"context"
	"time"

	"shared-models/agent"
	"shared-models/client"
	"shared-models/common"
	"shared-models/project"
)

// 示例：在 backend 服务中使用共享模型和客户端

// ProjectStageServiceExample 项目阶段服务示例
type ProjectStageServiceExample struct {
	agentClient *client.AgentClient
}

// NewProjectStageServiceExample 创建服务实例
func NewProjectStageServiceExample(agentsURL string) *ProjectStageServiceExample {
	agentClient := client.NewAgentClient(agentsURL, 5*time.Minute)
	return &ProjectStageServiceExample{
		agentClient: agentClient,
	}
}

// GeneratePRD 生成 PRD 示例
func (s *ProjectStageServiceExample) GeneratePRD(ctx context.Context, proj *project.Project) (*common.AgentResult, error) {
	req := &agent.GetPRDReq{
		ProjectGuid:  proj.GUID,
		Requirements: proj.Requirements,
	}

	return s.agentClient.GetPRD(ctx, req)
}

// DefineUXStandard 定义 UX 标准示例
func (s *ProjectStageServiceExample) DefineUXStandard(ctx context.Context, proj *project.Project) (*common.AgentResult, error) {
	req := &agent.GetUXStandardReq{
		ProjectGuid:  proj.GUID,
		Requirements: proj.Requirements,
		PrdPath:      "docs/PRD.md",
	}

	return s.agentClient.GetUXStandard(ctx, req)
}

// DesignArchitecture 设计架构示例
func (s *ProjectStageServiceExample) DesignArchitecture(ctx context.Context, proj *project.Project) (*common.AgentResult, error) {
	req := &agent.GetArchitectureReq{
		ProjectGuid:             proj.GUID,
		PrdPath:                 "docs/PRD.md",
		UxSpecPath:              "docs/ux/ux-spec.md",
		TemplateArchDescription: "Vue.js + Vite 前端，Go + Gin 后端，PostgreSQL 数据库，Redis 缓存，Docker 部署",
	}

	return s.agentClient.GetArchitecture(ctx, req)
}

// ImplementStory 实现故事示例
func (s *ProjectStageServiceExample) ImplementStory(ctx context.Context, proj *project.Project, epicFile, storyFile string) (*common.AgentResult, error) {
	req := &agent.ImplementStoryReq{
		ProjectGuid: proj.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch",
		DbFolder:    "docs/db",
		ApiFolder:   "docs/api",
		UxSpecPath:  "docs/ux/ux-spec.md",
		EpicFile:    epicFile,
		StoryFile:   storyFile,
	}

	return s.agentClient.ImplementStory(ctx, req)
}
