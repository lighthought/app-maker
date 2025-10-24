package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type PoHandler struct {
	agentTaskService services.AgentTaskService
}

func NewPoHandler(agentTaskService services.AgentTaskService) *PoHandler {
	return &PoHandler{agentTaskService: agentTaskService}
}

// GetEpicsAndStories godoc
// @Summary 获取史诗和用户故事
// @Description 基于PRD和架构设计生成Epics和Stories文档
// @Tags PO
// @Accept json
// @Produce json
// @Param request body agent.GetEpicsAndStoriesReq true "史诗故事请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/po/epicsandstories [get]
func (s *PoHandler) GetEpicsAndStories(c *gin.Context) {
	var req agent.GetEpicsAndStoriesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/po.md"
	} else {
		agentPrompt = "@bmad/po.mdc"
	}

	message := agentPrompt + " 我希望你基于PRD文档 @" + req.PrdPath + " 和 @" + req.ArchFolder +
		" 目录下的架构设计。首先创建分片的 Epics（史诗）和 Stories（用户故事），输出到 docs/stories/ 目录下。\n" +
		"注意：\n" +
		"1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 文件名必须使用英文命名，格式为: 'epic{N}-{english-name}-stories.md' 和 'epics.md'，不要使用中文文件名。例如 'epic1-project-creation-stories.md' 而不是 'epic1-项目创建-stories.md'。\n" +
		"3. 每个用户故事中要包含验收标准。不要考虑安全、合规。\n" +
		"4. 每个用户故事都要有自己的编号(如 US-001)，方便后续记录、跟踪。\n" +
		"5. 每个用户故事，要有完成情况勾选框，方便后续实现过程中更新进度。\n" +
		"6. 如果 docs/stories/ 目录下已经有完善的 Epics 和 Stories，直接返回概要信息，不用再尝试生成，原来的文档保持不变。\n" +
		"7. 在回答的最后，以 JSON 格式输出 MVP 阶段的 Epics 信息（通常是 P0 优先级的 Epics），格式如下:\n" +
		"```json\n" +
		"{\n" +
		"  \"mvp_epics\": [\n" +
		"    {\n" +
		"      \"epic_number\": 1,\n" +
		"      \"name\": \"Epic名称\",\n" +
		"      \"description\": \"Epic描述\",\n" +
		"      \"priority\": \"P0\",\n" +
		"      \"estimated_days\": 20,\n" +
		"      \"file_path\": \"docs/stories/epic1-xxx-stories.md\",\n" +
		"      \"stories\": [\n" +
		"        {\n" +
		"          \"story_number\": \"US-001\",\n" +
		"          \"title\": \"Story标题\",\n" +
		"          \"description\": \"Story描述\",\n" +
		"          \"priority\": \"P0\",\n" +
		"          \"estimated_days\": 3,\n" +
		"          \"depends\": \"依赖的其他Story\",\n" +
		"          \"techs\": \"技术要点\"\n" +
		"        }\n" +
		"      ]\n" +
		"    }\n" +
		"  ]\n" +
		"}\n" +
		"```\n"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypePO, message,
		req.CliTool, common.DevStatusPlanEpicAndStory)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取史诗和用户故事任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取史诗和用户故事成功", taskInfo.ID))
}
