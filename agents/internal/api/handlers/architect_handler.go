package handlers

import (
	"app-maker-agents/internal/services"
	"net/http"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

type ArchitectHandler struct {
	commandService *services.CommandService
}

func NewArchitectHandler(commandService *services.CommandService) *ArchitectHandler {
	return &ArchitectHandler{commandService: commandService}
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

	message := "@bmad/architect.mdc 请你基于最新的PRD文档 @" + req.PrdPath +
		" 和 UX 专家的设计文档 @" + req.UxSpecPath +
		" 帮我把整体架构设计 Architect.md, 前端架构设计 frontend_arch.md, 后端架构设计 backend_arch.md。" +
		" 都输出到 docs/arch/ 目录下。\n" +
		"注意：当前的项目代码是由模板生成，技术架构是：\n" + req.TemplateArchDescription
	/* TODO: 传入的地方从模板配置的地方读取
	"1. 前端：vue.js+ vite ；\n" +
	"2. 后端服务和 API： GO + Gin 框架实现 API、数据库用 PostgreSql、缓存用 Redis。\n" +
	"3. 部署相关的脚本已经有了，用的 docker，前端用一个 nginx ，配置 /api 重定向到 /backend:port ，这样就能在前端项目中访问后端 API 了。" +
	" 引用关系是：前端依赖后端，后端依赖 Redis 和 PostgreSql。"
	*/

	result := s.commandService.SimpleExecute(c.Request.Context(), req.ProjectGuid, message)
	if !result.Success {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "设计架构任务失败: "+result.Error))
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend

	c.JSON(http.StatusOK, utils.GetSuccessResponse("设计架构任务成功", result))
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

	message := "@bmad/architect.mdc 请你基于最新的PRD文档 @" + req.PrdPath +
		" 和 @" + req.ArchFolder + " 目录下的架构设计，以及 @" + req.StoriesFolder + " 目录下的用户故事，输出数据模型设计(可以用 sql 脚本代替)。" +
		"输出到 docs/db/ 目录下。"

	result := s.commandService.SimpleExecute(c.Request.Context(), req.ProjectGuid, message)
	if !result.Success {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "设计数据库任务失败: "+result.Error))
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend

	c.JSON(http.StatusOK, utils.GetSuccessResponse("设计数据库任务成功", result))
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

	message := "@bmad/architect.mdc 请你基于最新的PRD文档 @" + req.PrdPath +
		" 和 @" + req.DbFolder + " 目录下的数据模型，以及 @" + req.StoriesFolder + " 目录下的用户故事，生成 API 接口定义。" +
		" 输出到 docs/api/ 下多个文件（按控制器分类）。"

	result := s.commandService.SimpleExecute(c.Request.Context(), req.ProjectGuid, message)
	if !result.Success {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "设计 API 接口定义任务失败: "+result.Error))
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend

	c.JSON(http.StatusOK, utils.GetSuccessResponse("设计 API 接口定义任务成功", result))
}
