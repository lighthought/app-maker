package common

import (
	"shared-models/agent"
)

// 响应状态码
const (
	SUCCESS_CODE          = 0    // 成功
	ERROR_CODE            = 1    // 错误
	VALIDATION_ERROR      = 400  // 请求参数验证失败
	UNAUTHORIZED          = 401  // 未认证或认证失败
	FORBIDDEN             = 403  // 权限不足
	NOT_FOUND             = 404  // 资源不存在
	CONFLICT              = 409  // 资源冲突
	RATE_LIMIT_EXCEEDED   = 429  // 请求频率超限
	INTERNAL_ERROR        = 500  // 服务器内部错误
	SERVICE_UNAVAILABLE   = 503  // 服务不可用
	PROJECT_NOT_FOUND     = 2404 // 项目不存在
	PROJECT_ACCESS_DENIED = 2403 // 项目访问权限不足
	AGENT_SESSION_EXPIRED = 2410 // Agent会话已过期
	TASK_INTERNAL_ERROR   = 2500 // 任务内部错误
	DEPLOYMENT_ERROR      = 2501 // 部署错误
	INSUFFICIENT_QUOTA    = 2429 // 配额不足
)

// 通用状态
const (
	CommonStatusPending    = "pending"
	CommonStatusInProgress = "in_progress"
	CommonStatusDone       = "done"
	CommonStatusFailed     = "failed"
)

func GetProgressByCommonStatus(commandStatus string) int {
	switch commandStatus {
	case CommonStatusPending:
		return 0
	case CommonStatusInProgress:
		return 50
	case CommonStatusDone:
		return 100
	case CommonStatusFailed:
		return 0
	default:
		return 0
	}
}

type DevStatus string

// 开发阶段状态
const (
	DevStatusInitializing       = DevStatus("initializing")        // 等待开始
	DevStatusSetupEnvironment   = DevStatus("setup_environment")   // 环境处理
	DevStatusPendingAgents      = DevStatus("pending_agents")      // 等待Agents处理
	DevStatusCheckRequirement   = DevStatus("check_requirement")   // 需求检查
	DevStatusGeneratePRD        = DevStatus("generate_prd")        // 生成PRD
	DevStatusDefineUXStandard   = DevStatus("define_ux_standard")  // UX标准定义中
	DevStatusDesignArchitecture = DevStatus("design_architecture") // 架构设计中
	DevStatusPlanEpicAndStory   = DevStatus("plan_epic_and_story") // Epic和Story划分中
	DevStatusDefineDataModel    = DevStatus("define_data_model")   // 数据模型定义中
	DevStatusDefineAPI          = DevStatus("define_api")          // API接口定义中
	DevStatusDevelopStory       = DevStatus("develop_story")       // Story开发中
	DevStatusFixBug             = DevStatus("fix_bug")             // 问题修复中
	DevStatusRunTest            = DevStatus("run_test")            // 自动测试中
	DevStatusDeploy             = DevStatus("deploy")              // 部署中
	DevStatusDone               = DevStatus("done")                // 完成
	DevStatusFailed             = DevStatus("failed")              // 失败
)

// 获取开发阶段描述
func GetDevStageDescription(devStage DevStatus) string {
	switch devStage {
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
func GetDevStageProgress(devStage DevStatus) int {
	switch devStage {
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

// Agent 类型， 必须与 db_init.sql 中的 agent_role 一致
const (
	AgentTypeUser       = "user"
	AgentTypeAnalyse    = "analyst"
	AgentTypePM         = "pm"
	AgentTypeUX         = "ux-expert"
	AgentTypeArchitect  = "architect"
	AgentTypePO         = "po"
	AgentTypeDev        = "dev"
	AgentTypeQA         = "qa"
	AgentTypeSM         = "sm"
	AgentTypeBMADMaster = "bmad-master"
)

var (
	AgentAnalyst    = agent.Agent{Name: "Mary", Role: AgentTypeAnalyse, ChineseRole: "需求分析师"}
	AgentDev        = agent.Agent{Name: "James", Role: AgentTypeDev, ChineseRole: "开发工程师"}
	AgentPM         = agent.Agent{Name: "John", Role: AgentTypePM, ChineseRole: "产品经理"}
	AgentPO         = agent.Agent{Name: "Sarah", Role: AgentTypePO, ChineseRole: "产品负责人"}
	AgentArchitect  = agent.Agent{Name: "Winston", Role: AgentTypeArchitect, ChineseRole: "架构师"}
	AgentUXExpert   = agent.Agent{Name: "Sally", Role: AgentTypeUX, ChineseRole: "用户体验专家"}
	AgentQA         = agent.Agent{Name: "Quinn", Role: AgentTypeQA, ChineseRole: "测试和质量工程师"}
	AgentSM         = agent.Agent{Name: "Bob", Role: AgentTypeSM, ChineseRole: "敏捷教练"}
	AgentBMADMaster = agent.Agent{Name: "BMad Master", Role: AgentTypeBMADMaster, ChineseRole: "BMAD管理员"}
)

// 会话消息类型
const (
	ConversationTypeUser   = "user"
	ConversationTypeAgent  = "agent"
	ConversationTypeSystem = "system"
)

// WebSocket消息类型
const (
	WebSocketMessageTypePing                 = "ping"
	WebSocketMessageTypePong                 = "pong"
	WebSocketMessageTypeJoinProject          = "join_project"
	WebSocketMessageTypeLeaveProject         = "leave_project"
	WebSocketMessageTypeProjectStageUpdate   = "project_stage_update"
	WebSocketMessageTypeProjectMessage       = "project_message"
	WebSocketMessageTypeProjectInfoUpdate    = "project_info_update"
	WebSocketMessageTypeAgentMessage         = "agent_message"
	WebSocketMessageTypeUserFeedback         = "user_feedback"
	WebSocketMessageTypeUserFeedbackResponse = "user_feedback_response"
	WebSocketMessageTypeError                = "error"
)

// 任务类型常量
const (
	TypeProjectDownload    = "project:download"    // 下载项目
	TypeProjectBackup      = "project:backup"      // 备份项目
	TypeProjectInit        = "project:init"        // 初始化项目
	TypeProjectDevelopment = "project:development" // 开发项目
	TypeWebSocketBroadcast = "ws:broadcast"        // WebSocket 消息广播
)

// 任务优先级
const (
	TaskQueueCritical = 6 // 高优先级
	TaskQueueDefault  = 3 // 中优先级
	TaskQueueLow      = 1 // 低优先级
)
