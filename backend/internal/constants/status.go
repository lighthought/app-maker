package constants

import "shared-models/common"

func GetProgressByCommandStatus(commandStatus string) int {
	switch commandStatus {
	case common.CommonStatusPending:
		return 0
	case common.CommonStatusInProgress:
		return 50
	case common.CommonStatusDone:
		return 100
	case common.CommonStatusFailed:
		return 0
	default:
		return 0
	}
}

// 获取开发阶段描述
func GetDevStageDescription(devStage common.DevStage) string {
	switch devStage {
	case common.DevStatusInitializing:
		return "等待开始开发"
	case common.DevStatusSetupEnvironment:
		return "正在初始化开发环境"
	case common.DevStatusPendingAgents:
		return "等待Agents处理"
	case common.DevStatusCheckRequirement:
		return "正在检查需求"
	case common.DevStatusGeneratePRD:
		return "正在生成PRD文档"
	case common.DevStatusDefineUXStandard:
		return "正在定义UX标准"
	case common.DevStatusDesignArchitecture:
		return "正在设计系统架构"
	case common.DevStatusDefineDataModel:
		return "正在定义数据模型"
	case common.DevStatusDefineAPI:
		return "正在定义API接口"
	case common.DevStatusPlanEpicAndStory:
		return "正在划分Epic和Story"
	case common.DevStatusDevelopStory:
		return "正在开发Story功能"
	case common.DevStatusFixBug:
		return "正在修复开发问题"
	case common.DevStatusRunTest:
		return "正在执行自动测试"
	case common.DevStatusDeploy:
		return "正在部署项目"
	case common.DevStatusDone:
		return "项目开发完成"
	case common.DevStatusFailed:
		return "项目开发失败"
	default:
		return "未知状态"
	}
}

// 获取开发阶段进度
func GetDevStageProgress(devStage common.DevStage) int {
	switch devStage {
	case common.DevStatusInitializing:
		return 0
	case common.DevStatusSetupEnvironment:
		return 5
	case common.DevStatusPendingAgents:
		return 10
	case common.DevStatusCheckRequirement:
		return 15
	case common.DevStatusGeneratePRD:
		return 20
	case common.DevStatusDefineUXStandard:
		return 25
	case common.DevStatusDesignArchitecture:
		return 30
	case common.DevStatusDefineDataModel:
		return 35
	case common.DevStatusDefineAPI:
		return 40
	case common.DevStatusPlanEpicAndStory:
		return 45
	case common.DevStatusDevelopStory:
		return 60
	case common.DevStatusFixBug:
		return 75
	case common.DevStatusRunTest:
		return 90
	case common.DevStatusDeploy:
		return 95
	case common.DevStatusDone:
		return 100
	case common.DevStatusFailed:
		return 0
	default:
		return 0
	}
}
