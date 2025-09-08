package handlers

import (
	"net/http"
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
// @Param projectId path string true "项目ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(50)
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/chat/{projectId}/messages [get]
func (h *ChatHandler) GetProjectMessages(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
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

	// 获取对话消息
	messages, total, err := h.messageService.GetProjectConversations(c.Request.Context(), projectID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取对话历史失败"})
		return
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	hasNext := page < totalPages
	hasPrevious := page > 1

	c.JSON(http.StatusOK, gin.H{
		"code":        0,
		"message":     "success",
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
		"totalPages":  totalPages,
		"data":        messages,
		"hasNext":     hasNext,
		"hasPrevious": hasPrevious,
	})
}

// AddChatMessage 添加对话消息
// @Summary 添加对话消息
// @Description 为指定项目添加新的对话消息
// @Tags 对话消息
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param message body object true "对话消息" SchemaExample({"type":"user","content":"用户消息内容","isMarkdown":false})
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/chat/{projectId}/chat [post]
func (h *ChatHandler) AddChatMessage(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 创建对话消息
	message := &models.ConversationMessage{
		ProjectID:       projectID,
		Type:            req.Type,
		AgentRole:       req.AgentRole,
		AgentName:       req.AgentName,
		Content:         req.Content,
		IsMarkdown:      req.IsMarkdown,
		MarkdownContent: req.MarkdownContent,
		IsExpanded:      req.IsExpanded,
	}

	result, err := h.messageService.AddConversationMessage(c.Request.Context(), projectID, message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加对话消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}
