package constants

import "autocodeweb-backend/internal/models"

var (
	AgentAnalyst    = models.Agent{Name: "Mary", Role: "analyst", ChineseRole: "需求分析师"}
	AgentDev        = models.Agent{Name: "James", Role: "dev", ChineseRole: "开发工程师"}
	AgentPM         = models.Agent{Name: "John", Role: "pm", ChineseRole: "产品经理"}
	AgentPO         = models.Agent{Name: "Sarah", Role: "po", ChineseRole: "产品负责人"}
	AgentArchitect  = models.Agent{Name: "Winston", Role: "architect", ChineseRole: "架构师"}
	AgentUXExpert   = models.Agent{Name: "Sally", Role: "ux-expert", ChineseRole: "用户体验专家"}
	AgentQA         = models.Agent{Name: "Quinn", Role: "qa", ChineseRole: "测试和质量工程师"}
	AgentSM         = models.Agent{Name: "Bob", Role: "sm", ChineseRole: "敏捷教练"}
	AgentBMADMaster = models.Agent{Name: "BMad Master", Role: "bmad-master", ChineseRole: "BMAD管理员"}
)
