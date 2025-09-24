package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"autocodeweb-backend/internal/constants"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/tasks"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/cache"
	"autocodeweb-backend/pkg/logger"

	"github.com/hibiken/asynq"
)

// WebSocketService WebSocket 服务接口
type WebSocketService interface {
	// 连接管理
	RegisterClient(client *models.WebSocketClient)
	UnregisterClient(client *models.WebSocketClient)

	// 项目事件
	NotifyProjectStageUpdate(projectGUID string, stage *models.DevStage)
	NotifyProjectMessage(projectGUID string, message *models.ConversationMessage)
	NotifyProjectInfoUpdate(projectGUID string, info *models.Project)

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
	cache       cache.Cache
	cacheHour   time.Duration
}

// NewWebSocketService 创建 WebSocket 服务
func NewWebSocketService(asyncClient *asynq.Client, cache cache.Cache) WebSocketService {
	hub := models.NewWebSocketHub()
	cacheHour := 24 * time.Hour // TODO: 有了管理页面以后，根据配置页面里头的时间来设置

	return &webSocketService{
		hub:         hub,
		asyncClient: asyncClient,
		cache:       cache,
		cacheHour:   cacheHour,
	}
}

// saveDevStageToCache 保存开发阶段到缓存
func (s *webSocketService) saveDevStageToCache(projectGUID string, stage *models.DevStageInfo) error {
	var devStageCache models.DevStageCache
	var err error
	cacheKey := fmt.Sprintf("dev_stage:%s", projectGUID)

	if s.cache.Exists(cacheKey) {
		err = s.cache.Get(cacheKey, &devStageCache)
		if err != nil {
			return fmt.Errorf("获取开发阶段缓存失败: %w", err)
		}
	}

	devStageCache.Stages = append(devStageCache.Stages, *stage)
	devStageCache.ProjectGUID = projectGUID

	err = s.cache.Set(cacheKey, devStageCache, s.cacheHour)
	if err != nil {
		return fmt.Errorf("保存开发阶段缓存失败: %w", err)
	}

	return nil
}

// saveConversationMessageToCache 保存对话消息到缓存
func (s *webSocketService) saveConversationMessageToCache(projectGUID string, message *models.ConversationMessage) error {
	cacheKey := fmt.Sprintf("conversation_message:%s", projectGUID)
	var conversationMessageCache models.ConversationMessageCache
	var err error

	if s.cache.Exists(cacheKey) {
		err = s.cache.Get(cacheKey, &conversationMessageCache)
		if err != nil {
			return fmt.Errorf("获取对话消息缓存失败: %w", err)
		}
	}

	conversationMessageCache.Messages = append(conversationMessageCache.Messages, *message)
	conversationMessageCache.ProjectGUID = projectGUID

	err = s.cache.Set(cacheKey, conversationMessageCache, s.cacheHour)
	if err != nil {
		return fmt.Errorf("保存对话消息缓存失败: %w", err)
	}

	return nil
}

// SaveProjectInfoToCache 保存项目信息到缓存
func (s *webSocketService) saveProjectInfoUpdateToCache(projectGUID string, info *models.ProjectInfoUpdate) error {
	cacheKey := fmt.Sprintf("project_info_update:%s", projectGUID)
	var projectInfoCache models.ProjectInfoUpdate
	projectInfoCache.Copy(info)

	err := s.cache.Set(cacheKey, projectInfoCache, s.cacheHour)
	if err != nil {
		return fmt.Errorf("保存项目信息缓存失败: %w", err)
	}

	return nil
}

// getProjectStagesFromCache 从缓存中获取开发阶段
func (s *webSocketService) getProjectStagesFromCache(projectGUID string) (*[]models.DevStageInfo, error) {
	cacheKey := fmt.Sprintf("dev_stage:%s", projectGUID)

	if !s.cache.Exists(cacheKey) {
		logger.Info("开发阶段缓存不存在", logger.String("projectGUID", projectGUID))
		return nil, nil
	}

	var devStageCache models.DevStageCache
	err := s.cache.Get(cacheKey, &devStageCache)
	if err != nil {
		return nil, fmt.Errorf("获取开发阶段缓存失败: %w", err)
	}

	if len(devStageCache.Stages) == 0 {
		logger.Info("开发阶段缓存不存在", logger.String("projectGUID", projectGUID))
		return nil, nil
	}

	devStage := &devStageCache.Stages
	devStageCache.Stages = []models.DevStageInfo{}
	s.cache.Set(cacheKey, devStageCache, s.cacheHour)
	return devStage, nil
}

// getConversationMessagesFromCache 从缓存中获取对话消息
func (s *webSocketService) getConversationMessagesFromCache(projectGUID string) (*[]models.ConversationMessage, error) {
	cacheKey := fmt.Sprintf("conversation_message:%s", projectGUID)

	if !s.cache.Exists(cacheKey) {
		logger.Info("对话消息缓存不存在", logger.String("projectGUID", projectGUID))
		return nil, nil
	}

	var conversationMessageCache models.ConversationMessageCache
	err := s.cache.Get(cacheKey, &conversationMessageCache)
	if err != nil {
		return nil, fmt.Errorf("获取对话消息缓存失败: %w, projectGUID: %s", err, projectGUID)
	}

	if len(conversationMessageCache.Messages) == 0 {
		logger.Info("对话消息缓存不存在", logger.String("projectGUID", projectGUID))
		return nil, nil
	}

	messages := conversationMessageCache.Messages
	conversationMessageCache.Messages = []models.ConversationMessage{}
	s.cache.Set(cacheKey, conversationMessageCache, s.cacheHour)
	return &messages, nil
}

// GetProjectInfoUpdateFromCache 从缓存中获取项目更新信息
func (s *webSocketService) getProjectInfoUpdateFromCache(projectGUID string) (*models.ProjectInfoUpdate, error) {
	cacheKey := fmt.Sprintf("project_info_update:%s", projectGUID)
	var projectInfoCache models.ProjectInfoUpdate

	if !s.cache.Exists(cacheKey) {
		logger.Info("项目信息缓存不存在", logger.String("projectGUID", projectGUID))
		return nil, nil
	}

	err := s.cache.Get(cacheKey, &projectInfoCache)
	if err != nil {
		logger.Error("获取项目信息缓存失败", logger.String("error", err.Error()))
		return nil, fmt.Errorf("获取项目信息缓存失败: %w", err)
	}

	return &projectInfoCache, nil
}

// ProcessTask 处理任务
func (s *webSocketService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != models.TypeWebSocketBroadcast {
		return fmt.Errorf("不支持的任务类型: %s", task.Type())
	}

	payload := &models.WebSocketTaskPayload{}
	err := json.Unmarshal(task.Payload(), payload)
	if err != nil {
		logger.Error("解析WebSocket消息广播任务失败", logger.String("error", err.Error()))
		return err
	}

	projectGUID := payload.ProjectGUID
	messageType := payload.MessageType

	switch messageType {
	case constants.WebSocketMessageTypeProjectStageUpdate:
		devStages, err := s.getProjectStagesFromCache(projectGUID)
		if err != nil {
			logger.Error("获取开发阶段缓存失败", logger.String("error", err.Error()))
			return err
		}
		if devStages != nil {
			for _, devStage := range *devStages {
				s.broadcastProjectStage(projectGUID, &devStage)
			}
		}
		return nil
	case constants.WebSocketMessageTypeProjectMessage:
		messages, err := s.getConversationMessagesFromCache(projectGUID)
		if err != nil {
			logger.Error("获取对话消息缓存失败", logger.String("error", err.Error()))
			return err
		}
		if messages != nil {
			for _, message := range *messages {
				s.broadcastProjectMessage(projectGUID, &message)
			}
		}
		return nil
	case constants.WebSocketMessageTypeProjectInfoUpdate:
		projectInfoUpdate, err := s.getProjectInfoUpdateFromCache(projectGUID)
		if err != nil {
			logger.Error("获取项目信息缓存失败", logger.String("error", err.Error()))
			return err
		}
		if projectInfoUpdate != nil {
			s.broadcastProjectInfoUpdate(projectGUID, projectInfoUpdate)
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
func (s *webSocketService) broadcastProjectStage(projectGUID string, stage *models.DevStageInfo) {
	message := &models.WebSocketMessage{
		Type:        constants.WebSocketMessageTypeProjectStageUpdate,
		ProjectGUID: projectGUID,
		Data:        stage,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("stage_%s_%d", stage.ID, utils.GetTimeNow().Unix()),
	}
	s.broadcastToProject(projectGUID, message)
}

// broadcastProjectMessage 广播项目消息
func (s *webSocketService) broadcastProjectMessage(projectGUID string, message *models.ConversationMessage) {
	wsMessage := &models.WebSocketMessage{
		Type:        constants.WebSocketMessageTypeProjectMessage,
		ProjectGUID: projectGUID,
		Data:        message,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("message_%s_%d", message.ID, utils.GetTimeNow().Unix()),
	}
	s.broadcastToProject(projectGUID, wsMessage)
}

// broadcastProjectInfoUpdate 广播项目信息更新
func (s *webSocketService) broadcastProjectInfoUpdate(projectGUID string, info *models.ProjectInfoUpdate) {
	message := &models.WebSocketMessage{
		Type:        constants.WebSocketMessageTypeProjectInfoUpdate,
		ProjectGUID: projectGUID,
		Data:        info,
		Timestamp:   utils.GetCurrentTime(),
		ID:          fmt.Sprintf("info_%s_%d", projectGUID, utils.GetTimeNow().Unix()),
	}
	s.broadcastToProject(projectGUID, message)
}

// NotifyProjectStageUpdate 通知项目阶段更新
func (s *webSocketService) NotifyProjectStageUpdate(projectGUID string, stage *models.DevStage) {
	devStageInfo := &models.DevStageInfo{}
	devStageInfo.CopyFromDevStage(stage)

	err := s.saveDevStageToCache(projectGUID, devStageInfo)
	if err != nil {
		logger.Error("保存开发阶段缓存失败", logger.String("error", err.Error()))

		// 缓存获取失败，直接同步发送
		s.broadcastProjectStage(projectGUID, devStageInfo)
		return
	}

	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, constants.WebSocketMessageTypeProjectStageUpdate),
	)
	if err != nil {
		logger.Error("创建WebSocket消息广播任务失败", logger.String("error", err.Error()))

		// 异步失败，改为同步
		s.broadcastProjectStage(projectGUID, devStageInfo)
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
func (s *webSocketService) NotifyProjectMessage(projectGUID string, message *models.ConversationMessage) {
	err := s.saveConversationMessageToCache(projectGUID, message)
	if err != nil {
		logger.Error("保存对话消息缓存失败", logger.String("error", err.Error()))

		// 缓存获取失败，直接同步发送
		s.broadcastProjectMessage(projectGUID, message)
		return
	}

	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, constants.WebSocketMessageTypeProjectMessage),
	)
	if err != nil {
		logger.Error("创建WebSocket消息广播任务失败", logger.String("error", err.Error()))

		// 异步失败，改为同步
		s.broadcastProjectMessage(projectGUID, message)
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
func (s *webSocketService) NotifyProjectInfoUpdate(projectGUID string, project *models.Project) {
	info := project.GetUpdateInfo()
	err := s.saveProjectInfoUpdateToCache(projectGUID, info)
	if err != nil {
		logger.Error("保存项目信息缓存失败", logger.String("error", err.Error()))

		s.broadcastProjectInfoUpdate(projectGUID, info)
		return
	}

	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, constants.WebSocketMessageTypeProjectInfoUpdate),
	)
	if err != nil {
		logger.Error("创建WebSocket消息广播任务失败", logger.String("error", err.Error()))

		s.broadcastProjectInfoUpdate(projectGUID, info)
		return
	}

	logger.Info("项目信息更新通知已发送",
		logger.String("id", info.ID),
		logger.String("projectGUID", projectGUID),
		logger.String("name", info.Name),
		logger.String("type", constants.WebSocketMessageTypeProjectInfoUpdate),
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
