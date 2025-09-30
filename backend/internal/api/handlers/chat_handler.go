package handlers

import (
	"fmt"
	"net/http"
	"shared-models/common"
	"shared-models/utils"
	"strconv"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ChatHandler 对话消息处理器
type ChatHandler struct {
	messageService services.MessageService
	fileService    services.FileService
}

// NewChatHandler 创建对话消息处理器
func NewChatHandler(
	messageService services.MessageService,
	fileService services.FileService,
) *ChatHandler {
	return &ChatHandler{
		messageService: messageService,
		fileService:    fileService,
	}
}

// GetProjectMessages 获取项目对话历史
// @Summary 获取项目对话历史
// @Description 获取指定项目的对话消息历史记录，支持分页
// @Tags 对话消息
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(50)
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/chat/messages/{guid} [get]
func (h *ChatHandler) GetProjectMessages(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	// 添加调试日志
	fmt.Printf("DEBUG: ChatHandler - projectGuid=%s, page=%d, pageSize=%d, offset=%d\n",
		projectGuid, page, pageSize, offset)

	// 获取对话消息
	messages, total, err := h.messageService.GetProjectConversations(c.Request.Context(), projectGuid, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取对话历史失败, "+err.Error()))
		return
	}

	// 计算分页信息
	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrevious := page > 1

	c.JSON(http.StatusOK, models.PaginationResponse{
		Code:        common.SUCCESS_CODE,
		Message:     "success",
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Data:        messages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
		Timestamp:   utils.GetCurrentTime(),
	})
}

// AddChatMessage 添加对话消息
// @Summary 添加对话消息
// @Description 为指定项目添加新的对话消息
// @Tags 对话消息
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Param message body object true "对话消息" SchemaExample({"type":"user","content":"用户消息内容","isMarkdown":false})
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/chat/chat/{guid} [post]
func (h *ChatHandler) AddChatMessage(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	var req struct {
		Type            string `json:"type" binding:"required"`
		AgentRole       string `json:"agentRole,omitempty"`
		AgentName       string `json:"agentName,omitempty"`
		Content         string `json:"content"`
		IsMarkdown      bool   `json:"isMarkdown"`
		MarkdownContent string `json:"markdownContent,omitempty"`
		IsExpanded      bool   `json:"isExpanded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误"))
		return
	}

	if req.Type == "" {
		req.Type = common.ConversationTypeUser
	}
	if req.AgentRole == "" {
		req.AgentRole = common.AgentTypeUser
	}
	if req.AgentName == "" {
		req.AgentName = "user"
	}

	// 创建对话消息
	message := &models.ConversationMessage{
		ProjectGuid:     projectGuid,
		Type:            req.Type,
		AgentRole:       req.AgentRole,
		AgentName:       req.AgentName,
		Content:         req.Content,
		IsMarkdown:      req.IsMarkdown,
		MarkdownContent: req.MarkdownContent,
		IsExpanded:      req.IsExpanded,
	}

	result, err := h.messageService.AddConversationMessage(c.Request.Context(), message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "添加对话消息失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("添加对话消息成功", result))
}
