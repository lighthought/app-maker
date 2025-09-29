package handlers

import (
	"app-maker-agents/internal/services"
	"app-maker-agents/internal/utils"
	"net/http"
	"shared-models/agent"

	"time"

	"github.com/gin-gonic/gin"
)

type PoHandler struct {
	commandService *services.CommandService
}

func NewPoHandler(commandService *services.CommandService) *PoHandler {
	return &PoHandler{commandService: commandService}
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
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	message := "@bmad/po.mdc 我希望你基于PRD文档 @" + req.PrdPath + " 和 @" + req.ArchFolder +
		" 目录下的架构设计，创建 Epics（史诗）和 Stories（用户故事）。\n" +
		"生成分片的 Epics，输出到 docs/epics/ 下多个文件。再根据 Epics 生成分片的 Stories，输出到 docs/stories/ 下多个文件。" +
		"注意：stories 中要包含验收标准。不要考虑安全、合规。"

	result := s.commandService.Execute(c.Request.Context(), req.ProjectGuid, message, 5*time.Minute)
	if !result.Success {
		utils.Error(c, http.StatusInternalServerError, "分析任务失败: "+result.Error)
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend

	utils.Success(c, result.Output)
}
