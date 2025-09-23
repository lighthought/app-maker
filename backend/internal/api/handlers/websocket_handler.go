package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/auth"
	"autocodeweb-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHandler WebSocket 处理器
type WebSocketHandler struct {
	webSocketService services.WebSocketService
	projectService   services.ProjectService
	jwtService       *auth.JWTService
	upgrader         websocket.Upgrader
}

// NewWebSocketHandler 创建 WebSocket 处理器
func NewWebSocketHandler(
	webSocketService services.WebSocketService,
	projectService services.ProjectService,
	jwtService *auth.JWTService,
) *WebSocketHandler {
	return &WebSocketHandler{
		webSocketService: webSocketService,
		projectService:   projectService,
		jwtService:       jwtService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 在生产环境中应该检查 Origin
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// WebSocketUpgrade WebSocket 连接升级
func (h *WebSocketHandler) WebSocketUpgrade(c *gin.Context) {
	// 获取项目 GUID
	projectGUID := c.Param("guid")
	if projectGUID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "项目GUID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 获取 JWT Token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "缺少认证令牌",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 验证 JWT Token
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "认证令牌无效",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 检查项目访问权限
	project, err := h.projectService.CheckProjectAccess(c.Request.Context(), projectGUID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "无权限访问该项目",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 升级 HTTP 连接到 WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", logger.String("error", err.Error()))
		return
	}

	// 创建客户端连接
	client := &models.WebSocketClient{
		ID:          uuid.New().String(),
		UserID:      claims.UserID,
		ProjectGUID: projectGUID,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		LastPing:    utils.GetTimeNow(),
	}

	// 注册客户端
	h.webSocketService.RegisterClient(client)

	// 启动读写协程
	go h.writePump(client)
	go h.readPump(client)

	logger.Info("WebSocket 连接已建立",
		logger.String("clientID", client.ID),
		logger.String("userID", client.UserID),
		logger.String("projectGUID", projectGUID),
		logger.String("projectName", project.Name),
	)
}

// readPump 读取消息
func (h *WebSocketHandler) readPump(client *models.WebSocketClient) {
	defer func() {
		h.webSocketService.UnregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(utils.GetNSecondLater(60))
	client.Conn.SetPongHandler(func(string) error {
		client.LastPing = utils.GetTimeNow()
		client.Conn.SetReadDeadline(utils.GetNSecondLater(60))
		return nil
	})

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket 读取错误",
					logger.String("error", err.Error()),
					logger.String("clientID", client.ID),
				)
			}
			break
		}

		// 处理接收到的消息
		h.handleMessage(client, messageBytes)
	}
}

// writePump 写入消息
func (h *WebSocketHandler) writePump(client *models.WebSocketClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(utils.GetNSecondLater(10))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送消息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(utils.GetNSecondLater(10))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (h *WebSocketHandler) handleMessage(client *models.WebSocketClient, messageBytes []byte) {
	var message models.WebSocketMessage
	if err := json.Unmarshal(messageBytes, &message); err != nil {
		h.sendError(client, "消息格式错误", err.Error())
		return
	}

	switch message.Type {
	case "ping":
		h.handlePing(client)
	case "join_project":
		h.handleJoinProject(client, &message)
	case "leave_project":
		h.handleLeaveProject(client, &message)
	case "user_feedback":
		h.handleUserFeedback(client, &message)
	default:
		h.sendError(client, "未知消息类型", message.Type)
	}
}

// handlePing 处理心跳
func (h *WebSocketHandler) handlePing(client *models.WebSocketClient) {
	response := models.WebSocketMessage{
		Type:      "pong",
		Timestamp: utils.GetCurrentTime(),
		ID:        uuid.New().String(),
	}

	h.sendMessage(client, &response)
	client.LastPing = utils.GetTimeNow()
}

// handleJoinProject 处理加入项目
func (h *WebSocketHandler) handleJoinProject(client *models.WebSocketClient, message *models.WebSocketMessage) {
	// 验证项目访问权限
	_, err := h.projectService.CheckProjectAccess(context.Background(), message.ProjectGUID, client.UserID)
	if err != nil {
		h.sendError(client, "无权限访问该项目", err.Error())
		return
	}

	// 更新客户端的项目 GUID
	client.ProjectGUID = message.ProjectGUID

	response := models.WebSocketMessage{
		Type:        "project_joined",
		ProjectGUID: message.ProjectGUID,
		Data: map[string]string{
			"message": "成功加入项目",
		},
		Timestamp: utils.GetCurrentTime(),
		ID:        uuid.New().String(),
	}

	h.sendMessage(client, &response)

	logger.Info("客户端加入项目",
		logger.String("clientID", client.ID),
		logger.String("userID", client.UserID),
		logger.String("projectGUID", message.ProjectGUID),
	)
}

// handleLeaveProject 处理离开项目
func (h *WebSocketHandler) handleLeaveProject(client *models.WebSocketClient, message *models.WebSocketMessage) {
	response := models.WebSocketMessage{
		Type:        "project_left",
		ProjectGUID: message.ProjectGUID,
		Data: map[string]string{
			"message": "已离开项目",
		},
		Timestamp: utils.GetCurrentTime(),
		ID:        uuid.New().String(),
	}

	h.sendMessage(client, &response)

	logger.Info("客户端离开项目",
		logger.String("clientID", client.ID),
		logger.String("userID", client.UserID),
		logger.String("projectGUID", message.ProjectGUID),
	)
}

// handleUserFeedback 处理用户反馈
func (h *WebSocketHandler) handleUserFeedback(client *models.WebSocketClient, message *models.WebSocketMessage) {
	// TODO: 实现用户反馈处理逻辑
	// 这里可以转发给 agents-server 或保存到数据库

	response := models.WebSocketMessage{
		Type:        "user_feedback_response",
		ProjectGUID: message.ProjectGUID,
		Data: map[string]string{
			"message": "反馈已收到",
		},
		Timestamp: utils.GetCurrentTime(),
		ID:        uuid.New().String(),
	}

	h.sendMessage(client, &response)

	logger.Info("收到用户反馈",
		logger.String("clientID", client.ID),
		logger.String("userID", client.UserID),
		logger.String("projectGUID", message.ProjectGUID),
	)
}

// sendMessage 发送消息
func (h *WebSocketHandler) sendMessage(client *models.WebSocketClient, message *models.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("消息序列化失败", logger.String("error", err.Error()))
		return
	}

	select {
	case client.Send <- data:
	default:
		close(client.Send)
	}
}

// sendError 发送错误消息
func (h *WebSocketHandler) sendError(client *models.WebSocketClient, message, details string) {
	errorMessage := models.WebSocketMessage{
		Type: "error",
		Data: map[string]string{
			"message": message,
			"details": details,
		},
		Timestamp: utils.GetCurrentTime(),
		ID:        uuid.New().String(),
	}

	h.sendMessage(client, &errorMessage)
}

// GetWebSocketStats 获取 WebSocket 统计信息
func (h *WebSocketHandler) GetWebSocketStats(c *gin.Context) {
	stats := h.webSocketService.GetStats()

	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "获取 WebSocket 统计信息成功",
		Data:      stats,
		Timestamp: utils.GetCurrentTime(),
	})
}

// HealthCheck WebSocket 健康检查
func (h *WebSocketHandler) HealthCheck(c *gin.Context) {
	stats := h.webSocketService.GetStats()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": utils.GetCurrentTime(),
		"stats":     stats,
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "WebSocket 服务健康",
		Data:      health,
		Timestamp: utils.GetCurrentTime(),
	})
}
