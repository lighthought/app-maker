package common

import "shared-models/agent"

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

type DevStage string

// 开发阶段状态
const (
	DevStatusInitializing       = DevStage("initializing")        // 等待开始
	DevStatusSetupEnvironment   = DevStage("setup_environment")   // 环境处理
	DevStatusPendingAgents      = DevStage("pending_agents")      // 等待Agents处理
	DevStatusCheckRequirement   = DevStage("check_requirement")   // 需求检查
	DevStatusGeneratePRD        = DevStage("generate_prd")        // 生成PRD
	DevStatusDefineUXStandard   = DevStage("define_ux_standard")  // UX标准定义中
	DevStatusDesignArchitecture = DevStage("design_architecture") // 架构设计中
	DevStatusPlanEpicAndStory   = DevStage("plan_epic_and_story") // Epic和Story划分中
	DevStatusDefineDataModel    = DevStage("define_data_model")   // 数据模型定义中
	DevStatusDefineAPI          = DevStage("define_api")          // API接口定义中
	DevStatusDevelopStory       = DevStage("develop_story")       // Story开发中
	DevStatusFixBug             = DevStage("fix_bug")             // 问题修复中
	DevStatusRunTest            = DevStage("run_test")            // 自动测试中
	DevStatusDeploy             = DevStage("deploy")              // 部署中
	DevStatusDone               = DevStage("done")                // 完成
	DevStatusFailed             = DevStage("failed")              // 失败
)

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
