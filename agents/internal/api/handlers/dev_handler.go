package handlers

import (
	"app-maker-agents/internal/services"
	"net/http"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

type DevHandler struct {
	agentTaskService services.AgentTaskService
}

func NewDevHandler(agentTaskService services.AgentTaskService) *DevHandler {
	return &DevHandler{agentTaskService: agentTaskService}
}

// ImplementStory godoc
// @Summary 实现用户故事
// @Description 基于PRD、架构设计和UX标准实现用户故事
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.ImplementStoryReq true "实现故事请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/implstory [post]
func (h *DevHandler) ImplementStory(c *gin.Context) {
	var req agent.ImplementStoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	message := "@bmad/dev.mdc 请你始终记得项目的前后端框架及约束：\n" +
		"1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；\n" +
		"2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。\n" +
		"3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。\n" +
		"4. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程；\n" +
		"请你基于PRD文档 @" + req.PrdPath + " 和架构师的设计 @" + req.ArchFolder + " ，以及 UX 标准 @" + req.UxSpecPath +
		"  实现 @" + req.EpicFile + " 中的用户故事 @" + req.StoryFile + "。注意：\n" +
		"1. 数据库的设计在 @" + req.DbFolder + " 目录下。" + "API 的定义在 @" + req.ApiFolder + " 目录下。数据和接口如果在实现过程中需要调整，记得更新数据库设计和 API 定义文档" +
		"2. 实现完，编译确认下验收的标准是否都达到了，达到了以后，更新用户故事文档，勾上对应的验收标准。  \n" +
		"3. 然后再询问我，是否继续。不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。"

	taskInfo, err := h.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypeDev, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "实现用户故事任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("实现用户故事任务创建成功", taskInfo.ID))
}

// FixBug godoc
// @Summary 修复Bug
// @Description 根据Bug描述修复项目中的问题
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.FixBugReq true "修复Bug请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/fixbug [post]
func (h *DevHandler) FixBug(c *gin.Context) {
	var req agent.FixBugReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	message := "@bmad/dev.mdc 请你始终记得项目的前后端框架及约束：\n" +
		"1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；\n" +
		"2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。\n" +
		"3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。\n" +
		"4. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程。docs 目录下的架构、API、数据库和UX文档可以帮助你理解\n" +
		"5. 不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。" +
		"我当前遇到了 " + req.BugDescription + "，请你帮我修复下。"

	taskInfo, err := h.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypeDev, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "修复Bug任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("修复Bug任务创建成功", taskInfo.ID))
}

// RunTest godoc
// @Summary 运行测试
// @Description 执行项目的自动测试流程
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.FixBugReq true "运行测试请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/runtest [post]
func (h *DevHandler) RunTest(c *gin.Context) {
	var req agent.FixBugReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	message := "@bmad/dev.mdc 请你使用项目现有的测试脚本，完成项目的自动测试过程。包括前端的 lint 和后端的测试过程。\n" +
		"如果有 make test 命令，直接执行即可\n" +
		"注意：不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。"

	taskInfo, err := h.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypeDev, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "运行测试任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("测试任务创建成功", taskInfo.ID))
}

// Deploy godoc
// @Summary 部署项目
// @Description 执行项目的打包部署流程
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.FixBugReq true "部署请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/deploy [post]
func (h *DevHandler) Deploy(c *gin.Context) {
	var req agent.FixBugReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	message := "@bmad/dev.mdc 请你使用项目现有的打包脚本，完成项目的打包过程。\n" +
		"如果有类似 make build-dev 或 make build-prod 命令，直接执行即可。\n" +
		"注意：不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。"

	taskInfo, err := h.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypeDev, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "部署项目任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("部署项目任务创建成功", taskInfo.ID))
}
