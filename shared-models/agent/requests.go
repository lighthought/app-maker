package agent

import "encoding/json"

// 项目环境准备请求
type SetupProjEnvReq struct {
	ProjectGuid     string `json:"project_guid" binding:"required" example:"1234567890"`
	GitlabRepoUrl   string `json:"gitlab_repo_url" binding:"required" example:"https://gitlab.lighthought.com/app-maker/project-guid.git"`
	SetupBmadMethod bool   `json:"setup_bmad_method" binding:"required" example:"true"`
	BmadCliType     string `json:"bmad_cli_type" binding:"required" example:"claude"`
}

func (a *SetupProjEnvReq) ToBytes() []byte {
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil
	}
	return bytes
}

// 获取项目概览请求
type GetProjBriefReq struct {
	Requirements string `json:"requirements" binding:"required" example:"项目需求描述"`
	ProjectGuid  string `json:"project_guid" binding:"required" example:"1234567890"`
}

// 获取 PRD 请求
type GetPRDReq struct {
	ProjectGuid  string `json:"project_guid" binding:"required" example:"1234567890"`
	Requirements string `json:"requirements" binding:"required" example:"项目需求描述"`
}

// 获取 Epics 和 Stories 请求
type GetEpicsAndStoriesReq struct {
	ProjectGuid string `json:"project_guid" binding:"required" example:"1234567890"`
	PrdPath     string `json:"prd_path" binding:"required" example:"docs/PRD.md"`
	ArchFolder  string `json:"arch_folder" binding:"required" example:"docs/arch"`
}

// 获取 UX 标准请求
type GetUXStandardReq struct {
	ProjectGuid  string `json:"project_guid" binding:"required" example:"1234567890"`
	Requirements string `json:"requirements" binding:"required" example:"项目需求描述"`
	PrdPath      string `json:"prd_path" binding:"required" example:"docs/PRD.md"`
}

// 获取架构设计请求
type GetArchitectureReq struct {
	ProjectGuid             string `json:"project_guid" binding:"required" example:"1234567890"`
	PrdPath                 string `json:"prd_path" binding:"required" example:"docs/PRD.md"`
	UxSpecPath              string `json:"ux_spec_path" binding:"required" example:"docs/ux/ux-spec.md"`
	TemplateArchDescription string `json:"template_arch_description" binding:"required" example:"templates/architecture-template-v2.yaml"`
}

// 获取数据库设计请求
type GetDatabaseDesignReq struct {
	ProjectGuid   string `json:"project_guid" binding:"required" example:"1234567890"`
	PrdPath       string `json:"prd_path" binding:"required" example:"docs/PRD.md"`
	ArchFolder    string `json:"arch_folder" binding:"required" example:"docs/arch"`
	StoriesFolder string `json:"stories_folder" binding:"required" example:"docs/stories"`
}

// 获取 API 定义请求
type GetAPIDefinitionReq struct {
	ProjectGuid   string `json:"project_guid" binding:"required" example:"1234567890"`
	PrdPath       string `json:"prd_path" binding:"required" example:"docs/PRD.md"`
	DbFolder      string `json:"db_folder" binding:"required" example:"docs/db"`
	StoriesFolder string `json:"stories_folder" binding:"required" example:"docs/stories"`
}

// 实现用户故事请求
type ImplementStoryReq struct {
	ProjectGuid string `json:"project_guid" binding:"required" example:"1234567890"`
	PrdPath     string `json:"prd_path" binding:"required" example:"docs/PRD.md"`
	ArchFolder  string `json:"arch_folder" binding:"required" example:"docs/arch"`
	DbFolder    string `json:"db_folder" binding:"required" example:"docs/db"`
	ApiFolder   string `json:"api_folder" binding:"required" example:"docs/api"`
	UxSpecPath  string `json:"ux_spec_path" binding:"required" example:"docs/ux/ux-spec.md"`
	EpicFile    string `json:"epic_file" binding:"required" example:"docs/epics/epic.md"`
	StoryFile   string `json:"story_file" example:"docs/stories/story.md"`
}

// 修复 bug 请求
type FixBugReq struct {
	ProjectGuid    string `json:"project_guid" binding:"required" example:"1234567890"`
	BugDescription string `json:"bug_description" binding:"required" example:"bug description"`
}

// 运行测试请求
type RunTestReq struct {
	ProjectGuid string `json:"project_guid" validate:"required" example:"1234567890"`
}

// 部署请求
type DeployReq struct {
	ProjectGuid   string                 `json:"project_guid" validate:"required" example:"1234567890"`
	Environment   string                 `json:"environment,omitempty" example:"dev"` // dev, staging, prod
	DeployOptions map[string]interface{} `json:"deploy_options,omitempty"`
}
