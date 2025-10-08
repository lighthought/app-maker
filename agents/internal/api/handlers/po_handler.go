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
		" 目录下的架构设计。首先创建分片的 Epics（史诗）和 Stories（用户故事），输出到 docs/stories/ 目录下。\n" +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 文件名用史诗的名称命名，后缀和扩展名用 -story.md。\n" +
		"3. 每个用户故事中要包含验收标准。不要考虑安全、合规。\n" +
		"4. 每个用户故事都要有自己的编号，方便后续记录、跟踪。\n" +
		"5. 每个用户故事，预留完成情况勾选框，方便后续实现过程中更新进度。"

	taskInfo, err := s.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypePO, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取史诗和用户故事任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取史诗和用户故事成功", taskInfo.ID))
}
