package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type DevHandler struct {
	agentTaskService services.AgentTaskService
	commandService   services.CommandService
}

func NewDevHandler(agentTaskService services.AgentTaskService, commandService services.CommandService) *DevHandler {
	return &DevHandler{
		agentTaskService: agentTaskService,
		commandService:   commandService,
	}
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

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/dev.md"
	} else {
		agentPrompt = "@bmad/dev.mdc"
	}

	message := agentPrompt + " 请你基于PRD文档 @" + req.PrdPath + " 和架构师的设计 @" + req.ArchFolder + " ，以及 UX 标准 @" + req.UxSpecPath

	if req.StoryFile == "" {
		message += " 按照里程碑的顺序，实现 @" + req.EpicFile + " 中的下一个用户故事。\n"
	} else {
		message += " 按照里程碑的顺序，实现 @" + req.EpicFile + " 中的下一个用户故事 @" + req.StoryFile + "。\n"
	}

	message += "请你始终记得项目的前后端框架及约束：\n" +
		"1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；\n" +
		"2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。\n" +
		"3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。\n" +
		"注意：\n" +
		"1. 数据库的设计在 @" + req.DbFolder + " 目录下。" + "API 的定义在 @" + req.ApiFolder + " 目录下。数据和接口如果在实现过程中需要调整，记得更新数据库设计和 API 定义文档\n" +
		"2. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程；\n" +
		"3. 每次实现完，检查是否达成验收标准，更新对应 epic 的文档，勾上对应用户故事的验收标准。再更新 @" + req.EpicFile + " 中的 ReadMe.md 文件中的对应用户故事的完成状态。\n" +
		"4. 不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。\n" +
		"5. 实现过程中如果遇到问题，请自行尝试解决，解决不了再作为遗留问题输出到最后的总结中。\n" +
		"6. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"7. 每次实现完，记得修复编译问题，至少要保障项目能够 make build-dev 编译通过。"

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusDevelopStory)
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

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/dev.md"
	} else {
		agentPrompt = "@bmad/dev.mdc"
	}

	message := agentPrompt + " 我当前遇到了 " + req.BugDescription + "，请你帮我修复下。" +
		"请你始终记得项目的前后端框架及约束：\n" +
		"1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；\n" +
		"2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。\n" +
		"3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。\n\n" +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程。docs 目录下的架构、API、数据库和UX文档可以帮助你理解\n" +
		"3. 不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。"

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusFixBug)
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
	var req agent.RunTestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/dev.md"
	} else {
		agentPrompt = "@bmad/dev.mdc"
	}

	message := agentPrompt + " 请你使用项目现有的测试脚本，完成项目的自动测试过程。包括前端的 lint 和后端的测试过程。\n" +
		"如果有 make test 命令，直接执行即可\n" +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。"

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusRunTest)
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
	var req agent.DeployReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 异步执行部署任务
	taskInfo, err := h.agentTaskService.EnqueueDeployReq(&req)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "部署任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("部署任务已启动", taskInfo.ID))

}
