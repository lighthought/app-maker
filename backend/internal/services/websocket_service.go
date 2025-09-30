package services

import (
	"context"
	"encoding/json"
	"fmt"
	"shared-models/common"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"shared-models/logger"
	"shared-models/tasks"
	"shared-models/utils"

	"github.com/hibiken/asynq"
)

// WebSocketService WebSocket 服务接口
type WebSocketService interface {
	// 连接管理
	RegisterClient(client *models.WebSocketClient)
	UnregisterClient(client *models.WebSocketClient)

	// 项目事件
	NotifyProjectStageUpdate(ctx context.Context, projectGUID string, stage *models.DevStage)
	NotifyProjectMessage(ctx context.Context, projectGUID string, message *models.ConversationMessage)
	NotifyProjectInfoUpdate(ctx context.Context, projectGUID string, info *models.Project)

	// 启动和停止
	Start(ctx context.Context) error
	Stop() error

	// 健康检查
	GetStats() map[string]interface{}

	// ProcessTask 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// webSocketService WebSocket 服务实现
type webSocketService struct {
	hub         *models.WebSocketHub
	asyncClient *asynq.Client
	stageRepo   repositories.StageRepository
	messageRepo repositories.MessageRepository
	projectRepo repositories.ProjectRepository
}

// NewWebSocketService 创建 WebSocket 服务
func NewWebSocketService(asyncClient *asynq.Client,
	stageRepo repositories.StageRepository,
	messageRepo repositories.MessageRepository,
	projectRepo repositories.ProjectRepository,
) WebSocketService {
	return &webSocketService{
		hub:         models.NewWebSocketHub(),
		asyncClient: asyncClient,
		stageRepo:   stageRepo,
		messageRepo: messageRepo,
		projectRepo: projectRepo,
	}
}

// ProcessTask 处理任务
func (s *webSocketService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TypeWebSocketBroadcast {
		return fmt.Errorf("不支持的任务类型: %s", task.Type())
	}

	payload := &tasks.WebSocketTaskPayload{}
	err := json.Unmarshal(task.Payload(), payload)
	if err != nil {
		logger.Error("解析WebSocket消息广播任务失败", logger.String("error", err.Error()))
		return err
	}

	projectGUID := payload.ProjectGUID
	messageType := payload.MessageType

	switch messageType {
	case common.WebSocketMessageTypeProjectStageUpdate:
		if payload.StageID != "" {
			return s.broadcastProjectStage(ctx, projectGUID, payload.StageID)
		}
		return nil
	case common.WebSocketMessageTypeProjectMessage:
		if payload.MessageID != "" {
			return s.broadcastProjectMessage(ctx, projectGUID, payload.MessageID)
		}
		return nil
	case common.WebSocketMessageTypeProjectInfoUpdate:
		if payload.ProjectID != "" {
			return s.broadcastProjectInfoUpdate(ctx, projectGUID, payload.ProjectID)
		}
		return nil
	default:
		return fmt.Errorf("不支持的消息类型: %s", messageType)
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

// broadcastToProject 向指定项目广播消息
func (s *webSocketService) broadcastToProject(projectGUID string, message *models.WebSocketMessage) {
	message.ProjectGUID = projectGUID
	jsonMessage, _ := json.Marshal(message)
	logger.Info("[ws] 向指定项目广播消息",
		logger.String("projectGUID", projectGUID),
		logger.String("message", string(jsonMessage)))
	s.hub.Broadcast <- message
}

// broadcastToUser 向指定用户广播消息
func (s *webSocketService) broadcastToUser(userID string, message *models.WebSocketMessage) {
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

// broadcastToAll 向所有客户端广播消息
func (s *webSocketService) broadcastToAll(message *models.WebSocketMessage) {
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

// broadcastProjectStage 广播项目阶段
func (s *webSocketService) broadcastProjectStage(ctx context.Context, projectGUID, stageID string) error {
	if stageID == "" || projectGUID == "" {
		logger.Error("项目阶段ID或项目GUID为空")
		return fmt.Errorf("项目阶段ID或项目GUID为空")
	}

	stage, err := s.stageRepo.GetByID(ctx, stageID)
	if err != nil {
		logger.Error("获取项目阶段失败", logger.String("error", err.Error()))
		return fmt.Errorf("获取项目阶段失败: %w", err)
	}

	var stageInfo = models.DevStageInfo{}
	stageInfo.CopyFromDevStage(stage)
	message := &models.WebSocketMessage{
		Type:        common.WebSocketMessageTypeProjectStageUpdate,
		ProjectGUID: projectGUID,
		Data:        stageInfo,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("%s_%d", stage.ID, utils.GetTimeNow().Unix()),
	}
	s.broadcastToProject(projectGUID, message)
	return nil
}

// broadcastProjectMessage 广播项目消息
func (s *webSocketService) broadcastProjectMessage(ctx context.Context, projectGUID, messageID string) error {
	if messageID == "" || projectGUID == "" {
		logger.Error("项目消息ID或项目GUID为空")
		return fmt.Errorf("项目消息ID或项目GUID为空")
	}

	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		logger.Error("获取项目消息失败", logger.String("error", err.Error()))
		return fmt.Errorf("获取项目消息失败: %w", err)
	}
	wsMessage := &models.WebSocketMessage{
		Type:        common.WebSocketMessageTypeProjectMessage,
		ProjectGUID: projectGUID,
		Data:        message,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("%s_%d", message.ID, utils.GetTimeNow().Unix()),
	}
	s.broadcastToProject(projectGUID, wsMessage)
	return nil
}

// broadcastProjectInfoUpdate 广播项目信息更新
func (s *webSocketService) broadcastProjectInfoUpdate(ctx context.Context, projectGUID, projectID string) error {
	if projectID == "" || projectGUID == "" {
		logger.Error("项目ID或项目GUID为空")
		return fmt.Errorf("项目ID或项目GUID为空")
	}

	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		logger.Error("获取项目失败", logger.String("error", err.Error()))
		return fmt.Errorf("获取项目失败: %w", err)
	}

	projectInfo := project.GetUpdateInfo()

	message := &models.WebSocketMessage{
		Type:        common.WebSocketMessageTypeProjectInfoUpdate,
		ProjectGUID: projectGUID,
		Data:        projectInfo,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("%s_%d", projectInfo.ID, utils.GetTimeNow().Unix()),
	}
	s.broadcastToProject(projectGUID, message)
	return nil
}

// NotifyProjectStageUpdate 通知项目阶段更新
func (s *webSocketService) NotifyProjectStageUpdate(ctx context.Context, projectGUID string, stage *models.DevStage) {
	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, common.WebSocketMessageTypeProjectStageUpdate, stage.ID),
	)
	if err != nil {
		logger.Error("创建WebSocket消息广播任务失败", logger.String("error", err.Error()))

		// 异步失败，改为同步
		s.broadcastProjectStage(ctx, projectGUID, stage.ID)
		return
	}

	logger.Info("项目阶段更新通知异步发送",
		logger.String("projectGUID", projectGUID),
		logger.String("stageID", stage.ID),
		logger.String("stageName", stage.Name),
		logger.String("status", stage.Status),
		logger.String("taskID", taskInfo.ID),
	)
}

// NotifyProjectMessage 通知项目新消息
func (s *webSocketService) NotifyProjectMessage(ctx context.Context, projectGUID string, message *models.ConversationMessage) {
	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, common.WebSocketMessageTypeProjectMessage, message.ID),
	)
	if err != nil {
		logger.Error("创建WebSocket消息广播任务失败", logger.String("error", err.Error()))

		// 异步失败，改为同步
		s.broadcastProjectMessage(ctx, projectGUID, message.ID)
		return
	}

	logger.Info("项目消息通知异步发送",
		logger.String("projectGUID", projectGUID),
		logger.String("messageID", message.ID),
		logger.String("messageType", message.Type),
		logger.String("taskID", taskInfo.ID),
	)
}

// NotifyProjectInfoUpdate 通知项目信息更新
func (s *webSocketService) NotifyProjectInfoUpdate(ctx context.Context, projectGUID string, project *models.Project) {
	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, common.WebSocketMessageTypeProjectInfoUpdate, project.ID),
	)
	if err != nil {
		logger.Error("创建WebSocket消息广播任务失败", logger.String("error", err.Error()))

		s.broadcastProjectInfoUpdate(ctx, projectGUID, project.ID)
		return
	}

	logger.Info("项目信息更新通知异步发送",
		logger.String("id", project.ID),
		logger.String("projectGUID", projectGUID),
		logger.String("name", project.Name),
		logger.String("type", common.WebSocketMessageTypeProjectInfoUpdate),
		logger.String("taskID", taskInfo.ID),
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
