package handlers

import (
	"app-maker-agents/internal/services"
	"net/http"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"

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

	message := "@bmad/po.mdc 我希望你基于PRD文档 @" + req.PrdPath + " 和 @" + req.ArchFolder +
		" 目录下的架构设计，创建 Epics（史诗）和 Stories（用户故事）。\n" +
		"生成 Epics，输出到 docs/ 目录下，文件名为： epics-stories.md。\n" +
		"注意：1. 输出分片的 Stories，输出到 docs/stories/ 目录下。文件名用 epic 的名称命名，后缀用 -story.md。\n" +
		"2. stories 中要包含验收标准。不要考虑安全、合规。每个用户故事都要有自己的编号，方便后续记录、跟踪。"

	taskInfo, err := s.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypePO, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取史诗和用户故事任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取史诗和用户故事成功", taskInfo.ID))
}
