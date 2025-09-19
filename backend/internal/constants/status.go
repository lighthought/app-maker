package constants

const (
	CommandStatusPending    = "pending"
	CommandStatusInProgress = "in_progress"
	CommandStatusDone       = "done"
	CommandStatusFailed     = "failed"
)

func GetProgressByCommandStatus(commandStatus string) int {
	switch commandStatus {
	case CommandStatusPending:
		return 0
	case CommandStatusInProgress:
		return 50
	case CommandStatusDone:
		return 100
	case CommandStatusFailed:
		return 0
	default:
		return 0
	}
}

// 开发子状态常量
const (
	DevStatusInitializing       = "initializing"        // 等待开始
	DevStatusSetupEnvironment   = "setup_environment"   // 环境处理
	DevStatusPendingAgents      = "pending_agents"      // 等待Agents处理
	DevStatusCheckRequirement   = "check_requirement"   // 需求检查
	DevStatusGeneratePRD        = "generate_prd"        // 生成PRD
	DevStatusDefineUXStandard   = "define_ux_standard"  // UX标准定义中
	DevStatusDesignArchitecture = "design_architecture" // 架构设计中
	DevStatusDefineDataModel    = "define_data_model"   // 数据模型定义中
	DevStatusDefineAPI          = "define_api"          // API接口定义中
	DevStatusPlanEpicAndStory   = "plan_epic_and_story" // Epic和Story划分中
	DevStatusDevelopStory       = "develop_story"       // Story开发中
	DevStatusFixBug             = "fix_bug"             // 问题修复中
	DevStatusRunTest            = "run_test"            // 自动测试中
	DevStatusDeploy             = "deploy"              // 部署中
	DevStatusDone               = "done"                // 完成
	DevStatusFailed             = "failed"              // 失败
)

// 获取开发阶段描述
func GetDevStageDescription(devStatus string) string {
	switch devStatus {
	case DevStatusInitializing:
		return "等待开始开发"
	case DevStatusSetupEnvironment:
		return "正在初始化开发环境"
	case DevStatusPendingAgents:
		return "等待Agents处理"
	case DevStatusCheckRequirement:
		return "正在检查需求"
	case DevStatusGeneratePRD:
		return "正在生成PRD文档"
	case DevStatusDefineUXStandard:
		return "正在定义UX标准"
	case DevStatusDesignArchitecture:
		return "正在设计系统架构"
	case DevStatusDefineDataModel:
		return "正在定义数据模型"
	case DevStatusDefineAPI:
		return "正在定义API接口"
	case DevStatusPlanEpicAndStory:
		return "正在划分Epic和Story"
	case DevStatusDevelopStory:
		return "正在开发Story功能"
	case DevStatusFixBug:
		return "正在修复开发问题"
	case DevStatusRunTest:
		return "正在执行自动测试"
	case DevStatusDeploy:
		return "正在部署项目"
	case DevStatusDone:
		return "项目开发完成"
	case DevStatusFailed:
		return "项目开发失败"
	default:
		return "未知状态"
	}
}

// 获取开发阶段进度
func GetDevStageProgress(devStatus string) int {
	switch devStatus {
	case DevStatusInitializing:
		return 0
	case DevStatusSetupEnvironment:
		return 5
	case DevStatusPendingAgents:
		return 10
	case DevStatusCheckRequirement:
		return 15
	case DevStatusGeneratePRD:
		return 20
	case DevStatusDefineUXStandard:
		return 25
	case DevStatusDesignArchitecture:
		return 30
	case DevStatusDefineDataModel:
		return 35
	case DevStatusDefineAPI:
		return 40
	case DevStatusPlanEpicAndStory:
		return 45
	case DevStatusDevelopStory:
		return 60
	case DevStatusFixBug:
		return 75
	case DevStatusRunTest:
		return 90
	case DevStatusDeploy:
		return 95
	case DevStatusDone:
		return 100
	case DevStatusFailed:
		return 0
	default:
		return 0
	}
}
