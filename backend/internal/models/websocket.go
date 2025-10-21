package models

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage WebSocket 消息结构
type WebSocketMessage struct {
	Type        string      `json:"type"`
	ProjectGUID string      `json:"projectGuid"`
	Data        interface{} `json:"data"`
	Timestamp   string      `json:"timestamp"`
	ID          string      `json:"id"`
}

// WebSocketClient WebSocket 客户端连接
type WebSocketClient struct {
	ID          string
	UserID      string
	ProjectGUID string
	Conn        *websocket.Conn
	Send        chan []byte
	Hub         *WebSocketHub
	LastPing    time.Time
}

// WebSocketHub WebSocket 连接管理器
type WebSocketHub struct {
	Clients    map[*WebSocketClient]bool
	Projects   map[string]map[*WebSocketClient]bool
	Register   chan *WebSocketClient
	Unregister chan *WebSocketClient
	Broadcast  chan *WebSocketMessage
	Mutex      sync.RWMutex
}

// NewWebSocketHub 创建新的 WebSocket Hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		Clients:    make(map[*WebSocketClient]bool),
		Projects:   make(map[string]map[*WebSocketClient]bool),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
		Broadcast:  make(chan *WebSocketMessage, 256),
	}
}

// RegisterClient 注册客户端连接
func (h *WebSocketHub) RegisterClient(client *WebSocketClient) {
	h.Mutex.Lock()
	h.Clients[client] = true
	if h.Projects[client.ProjectGUID] == nil {
		h.Projects[client.ProjectGUID] = make(map[*WebSocketClient]bool)
	}
	h.Projects[client.ProjectGUID][client] = true
	h.Mutex.Unlock()
}

// UnregisterClient 注销客户端连接
func (h *WebSocketHub) UnregisterClient(client *WebSocketClient) {
	h.Mutex.Lock()
	if _, ok := h.Clients[client]; ok {
		delete(h.Clients, client)
		close(client.Send)

		if projectClients, exists := h.Projects[client.ProjectGUID]; exists {
			delete(projectClients, client)
			if len(projectClients) == 0 {
				delete(h.Projects, client.ProjectGUID)
			}
		}
	}
	h.Mutex.Unlock()
}

// 处理客户端消息
func (h *WebSocketHub) HandleClientMessage(message *WebSocketMessage) {
	h.Mutex.RLock()
	if projectClients, exists := h.Projects[message.ProjectGUID]; exists {
		// 将整个消息序列化为 JSON 字节数组
		var dataBytes []byte
		if jsonData, err := json.Marshal(message); err == nil {
			dataBytes = jsonData
		} else {
			// 如果序列化失败，发送错误消息
			errorMsg := map[string]interface{}{
				"type":      "error",
				"message":   "Failed to serialize WebSocket message",
				"timestamp": time.Now().Format(time.RFC3339),
			}
			if errorData, err := json.Marshal(errorMsg); err == nil {
				dataBytes = errorData
			}
		}

		for client := range projectClients {
			select {
			case client.Send <- dataBytes:
			default:
				close(client.Send)
				delete(h.Clients, client)
				delete(projectClients, client)
			}
		}
	}
	h.Mutex.RUnlock()
}

// Run 运行 Hub 主循环
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.RegisterClient(client)

		case client := <-h.Unregister:
			h.UnregisterClient(client)

		case message := <-h.Broadcast:
			h.HandleClientMessage(message)
		}
	}
}
