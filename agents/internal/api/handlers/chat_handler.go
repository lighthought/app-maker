package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	agentTaskService services.AgentTaskService
}

func NewChatHandler(agentTaskService services.AgentTaskService) *ChatHandler {
	return &ChatHandler{
		agentTaskService: agentTaskService,
	}
}

// ChatWithAgent 与指定 Agent 对话
// @Summary 与指定 Agent 对话
// @Description 向指定 Agent 发送消息，使用现有会话继续对话
// @Tags Chat
// @Accept json
// @Produce json
// @Param request body agent.ChatReq true "对话请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/chat [post]
func (h *ChatHandler) ChatWithAgent(c *gin.Context) {
	var req agent.ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 调用 agentTaskService 的 ChatWithAgent 方法
	taskInfo, err := h.agentTaskService.EnqueueChatWithAgent(&req)

	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("创建 Agent 对话任务成功", taskInfo.ID))
}
