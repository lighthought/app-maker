package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type ArchitectHandler struct {
	agentTaskService services.AgentTaskService
}

func NewArchitectHandler(agentTaskService services.AgentTaskService) *ArchitectHandler {
	return &ArchitectHandler{agentTaskService: agentTaskService}
}

// GetArchitecture godoc
// @Summary 获取架构设计
// @Description 基于PRD和UX设计生成整体架构、前端架构和后端架构设计文档
// @Tags Architect
// @Accept json
// @Produce json
// @Param request body agent.GetArchitectureReq true "架构设计请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/architect/architect [get]
func (s *ArchitectHandler) GetArchitecture(c *gin.Context) {
	var req agent.GetArchitectureReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/architect.md"
	} else {
		agentPrompt = "@bmad/architect.mdc"
	}

	message := agentPrompt + " 请你基于最新的PRD文档 @" + req.PrdPath +
		" 和 UX 专家的设计文档 @" + req.UxSpecPath +
		" 帮我把整体架构设计 Architect.md, 前端架构设计 frontend_arch.md, 后端架构设计 backend_arch.md。" +
		" 都输出到 docs/arch/ 目录下。\n" +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 重要: 所有生成的文件名必须使用英文命名，不要使用中文文件名。\n" +
		"3. 当前的项目代码是由模板生成，所以当前可能存在一些不在 PRD 描述内的实现细节，不影响编译可以不考虑。\n" +
		"4. 当前项目使用的模板技术架构是：\n" + req.TemplateArchDescription +
		"5. 如果 docs/arch/ 目录下已经有完善的架构设计，直接返回概要信息，不用再尝试生成，原来的文档保持不变。"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeArchitect, message,
		req.CliTool, common.DevStatusDesignArchitecture)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "异步任务压入失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("设计架构任务创建成功", taskInfo.ID))
}

// GetDatabaseDesign godoc
// @Summary 获取数据库设计
// @Description 基于PRD、架构设计和用户故事生成数据模型设计
// @Tags Architect
// @Accept json
// @Produce json
// @Param request body agent.GetDatabaseDesignReq true "数据库设计请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/architect/database [get]
func (s *ArchitectHandler) GetDatabaseDesign(c *gin.Context) {
	var req agent.GetDatabaseDesignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/architect.md"
	} else {
		agentPrompt = "@bmad/architect.mdc"
	}

	message := agentPrompt + " 请你基于最新的PRD文档 @" + req.PrdPath +
		" 和 @" + req.ArchFolder + " 目录下的架构设计，以及 @" + req.StoriesFolder +
		" 目录下的用户故事，输出数据模型设计(可以用 sql 脚本代替)。输出到 docs/db/ 目录下。\n" +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 重要: 所有生成的文件名必须使用英文命名，不要使用中文文件名。\n" +
		"3. 如果 docs/db/ 目录下已经有完善的数据模型设计，直接返回概要信息，不用再尝试生成，原来的文档保持不变。"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeArchitect, message,
		req.CliTool, common.DevStatusDefineDataModel)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "设计数据库任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("设计数据库任务成功", taskInfo.ID))
}

// GetAPIDefinition godoc
// @Summary 获取API定义
// @Description 基于PRD、数据模型和用户故事生成API接口定义
// @Tags Architect
// @Accept json
// @Produce json
// @Param request body agent.GetAPIDefinitionReq true "API定义请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/architect/apidefinition [get]
func (s *ArchitectHandler) GetAPIDefinition(c *gin.Context) {
	var req agent.GetAPIDefinitionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/architect.md"
	} else {
		agentPrompt = "@bmad/architect.mdc"
	}

	message := agentPrompt + " 请你基于最新的PRD文档 @" + req.PrdPath +
		" 和 @" + req.DbFolder + " 目录下的数据模型，以及 @" + req.StoriesFolder + " 目录下的用户故事，生成 API 接口定义。输出到 docs/api/ 下多个文件（按控制器分类）。\n" +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 重要: 所有生成的文件名必须使用英文命名，不要使用中文文件名。\n" +
		"3. 如果 docs/api/ 目录下已经有完善的 API 接口定义，直接返回概要信息，不用再尝试生成，原来的文档保持不变。"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeArchitect, message,
		req.CliTool, common.DevStatusDefineAPI)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "设计 API 接口定义任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("设计 API 接口定义任务成功", taskInfo.ID))
}
