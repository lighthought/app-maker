package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"
)

// WebSocketService WebSocket 服务接口
type WebSocketService interface {
	// 连接管理
	RegisterClient(client *models.WebSocketClient)
	UnregisterClient(client *models.WebSocketClient)

	// 消息广播
	BroadcastToProject(projectGUID string, message *models.WebSocketMessage)
	BroadcastToUser(userID string, message *models.WebSocketMessage)
	BroadcastToAll(message *models.WebSocketMessage)

	// 项目事件
	NotifyProjectStageUpdate(projectGUID string, stage *models.DevStage)
	NotifyProjectMessage(projectGUID string, message *models.ConversationMessage)
	NotifyProjectInfoUpdate(projectGUID string, info *models.Project)

	// 启动和停止
	Start(ctx context.Context) error
	Stop() error

	// 健康检查
	GetStats() map[string]interface{}
}

// webSocketService WebSocket 服务实现
type webSocketService struct {
	hub *models.WebSocketHub
}

// NewWebSocketService 创建 WebSocket 服务
func NewWebSocketService() WebSocketService {
	hub := models.NewWebSocketHub()

	return &webSocketService{
		hub: hub,
	}
}

// RegisterClient 注册客户端连接
func (s *webSocketService) RegisterClient(client *models.WebSocketClient) {
	logger.Info("[ws] 注册客户端连接",
		logger.String("clientID", client.ID),
		logger.String("userID", client.UserID),
		logger.String("projectGUID", client.ProjectGUID))
	s.hub.Register <- client
}

// UnregisterClient 注销客户端连接
func (s *webSocketService) UnregisterClient(client *models.WebSocketClient) {
	logger.Info("[ws] 注销客户端连接",
		logger.String("clientID", client.ID),
		logger.String("userID", client.UserID),
		logger.String("projectGUID", client.ProjectGUID))
	s.hub.Unregister <- client
}

// BroadcastToProject 向指定项目广播消息
func (s *webSocketService) BroadcastToProject(projectGUID string, message *models.WebSocketMessage) {
	message.ProjectGUID = projectGUID
	jsonMessage, _ := json.Marshal(message)
	logger.Info("[ws] 向指定项目广播消息",
		logger.String("projectGUID", projectGUID),
		logger.String("message", string(jsonMessage)))
	s.hub.Broadcast <- message
}

// BroadcastToUser 向指定用户广播消息
func (s *webSocketService) BroadcastToUser(userID string, message *models.WebSocketMessage) {
	s.hub.Mutex.RLock()
	defer s.hub.Mutex.RUnlock()
	jsonMessage, _ := json.Marshal(message)
	logger.Info("[ws] 向指定用户广播消息",
		logger.String("userID", userID),
		logger.String("message", string(jsonMessage)))

	for client := range s.hub.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- s.serializeMessage(message):
			default:
				close(client.Send)
				delete(s.hub.Clients, client)
				if projectClients, exists := s.hub.Projects[client.ProjectGUID]; exists {
					delete(projectClients, client)
					if len(projectClients) == 0 {
						delete(s.hub.Projects, client.ProjectGUID)
					}
				}
			}
		}
	}
}

// BroadcastToAll 向所有客户端广播消息
func (s *webSocketService) BroadcastToAll(message *models.WebSocketMessage) {
	s.hub.Mutex.RLock()
	defer s.hub.Mutex.RUnlock()
	jsonMessage, _ := json.Marshal(message)
	logger.Info("[ws] 向所有客户端广播消息",
		logger.String("message", string(jsonMessage)))

	for client := range s.hub.Clients {
		select {
		case client.Send <- s.serializeMessage(message):
		default:
			close(client.Send)
			delete(s.hub.Clients, client)
			if projectClients, exists := s.hub.Projects[client.ProjectGUID]; exists {
				delete(projectClients, client)
				if len(projectClients) == 0 {
					delete(s.hub.Projects, client.ProjectGUID)
				}
			}
		}
	}
}

// NotifyProjectStageUpdate 通知项目阶段更新
func (s *webSocketService) NotifyProjectStageUpdate(projectGUID string, stage *models.DevStage) {
	message := &models.WebSocketMessage{
		Type:        "project_stage_update",
		ProjectGUID: projectGUID,
		Data:        stage,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("stage_%s_%d", stage.ID, utils.GetTimeNow().Unix()),
	}

	s.BroadcastToProject(projectGUID, message)

	logger.Info("项目阶段更新通知已发送",
		logger.String("projectGUID", projectGUID),
		logger.String("stageID", stage.ID),
		logger.String("stageName", stage.Name),
		logger.String("status", stage.Status),
	)
}

// NotifyProjectMessage 通知项目新消息
func (s *webSocketService) NotifyProjectMessage(projectGUID string, message *models.ConversationMessage) {
	wsMessage := &models.WebSocketMessage{
		Type:        "project_message",
		ProjectGUID: projectGUID,
		Data:        message,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("message_%s_%d", message.ID, utils.GetTimeNow().Unix()),
	}

	s.BroadcastToProject(projectGUID, wsMessage)

	logger.Info("项目消息通知已发送",
		logger.String("projectGUID", projectGUID),
		logger.String("messageID", message.ID),
		logger.String("messageType", message.Type),
	)
}

// NotifyProjectInfoUpdate 通知项目信息更新
func (s *webSocketService) NotifyProjectInfoUpdate(projectGUID string, project *models.Project) {
	info := models.ConvertToProjectInfo(project)
	message := &models.WebSocketMessage{
		Type:        "project_info_update",
		ProjectGUID: projectGUID,
		Data:        info,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("info_%s_%d", projectGUID, utils.GetTimeNow().Unix()),
	}

	s.BroadcastToProject(projectGUID, message)

	logger.Info("项目信息更新通知已发送",
		logger.String("projectGUID", message.ProjectGUID),
		logger.String("timestamp", message.Timestamp),
		logger.String("id", message.ID),
		logger.String("type", message.Type),
	)
}

// Start 启动 WebSocket 服务
func (s *webSocketService) Start(ctx context.Context) error {
	logger.Info("WebSocket 服务启动中...")

	go s.hub.Run()

	// 启动心跳检测
	go s.startHeartbeat(ctx)

	logger.Info("WebSocket 服务启动完成")
	return nil
}

// Stop 停止 WebSocket 服务
func (s *webSocketService) Stop() error {
	logger.Info("WebSocket 服务停止中...")

	// 关闭所有客户端连接
	s.hub.Mutex.Lock()
	for client := range s.hub.Clients {
		client.Conn.Close()
		close(client.Send)
	}
	s.hub.Mutex.Unlock()

	logger.Info("WebSocket 服务已停止")
	return nil
}

// GetStats 获取服务统计信息
func (s *webSocketService) GetStats() map[string]interface{} {
	s.hub.Mutex.RLock()
	defer s.hub.Mutex.RUnlock()

	stats := map[string]interface{}{
		"total_clients":   len(s.hub.Clients),
		"total_projects":  len(s.hub.Projects),
		"active_projects": make(map[string]int),
	}

	for projectGUID, clients := range s.hub.Projects {
		stats["active_projects"].(map[string]int)[projectGUID] = len(clients)
	}

	return stats
}

// serializeMessage 序列化消息
func (s *webSocketService) serializeMessage(message *models.WebSocketMessage) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("消息序列化失败", logger.String("error", err.Error()))
		return nil
	}
	return data
}

// startHeartbeat 启动心跳检测
func (s *webSocketService) startHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.hub.Mutex.RLock()
			for client := range s.hub.Clients {
				if time.Since(client.LastPing) > 60*time.Second {
					logger.Warn("客户端心跳超时，断开连接",
						logger.String("clientID", client.ID),
						logger.String("userID", client.UserID),
					)
					client.Conn.Close()
				}
			}
			s.hub.Mutex.RUnlock()
		}
	}
}
