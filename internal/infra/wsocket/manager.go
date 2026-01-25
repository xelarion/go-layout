package wsocket

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// HandlerFunc 消息处理函数
type HandlerFunc func(ctx *Context)

// Manager 管理所有 WebSocket 连接
type Manager struct {
	// 源ID -> 连接集合（一个源可能多个连接）
	clients map[string]map[*Client]struct{}
	mu      sync.RWMutex

	register   chan *Client
	unregister chan *Client

	handlers   map[string]HandlerFunc
	handlersMu sync.RWMutex

	upgrader websocket.Upgrader
	logger   *zap.Logger
}

// NewManager 创建连接管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		clients:    make(map[string]map[*Client]struct{}),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		handlers:   make(map[string]HandlerFunc),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		logger: logger.Named("wsocket"),
	}
}

// Run 启动事件循环（需在 goroutine 中调用）
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.addClient(client)
		case client := <-m.unregister:
			m.removeClient(client)
		}
	}
}

// Upgrade 升级 HTTP 连接为 WebSocket 并注册
func (m *Manager) Upgrade(w http.ResponseWriter, r *http.Request, sourceID string) (*Client, error) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	client := newClient(m, conn, sourceID)
	m.register <- client

	go client.writePump()
	go client.readPump()

	return client, nil
}

// Handle 注册消息处理器
func (m *Manager) Handle(action string, handler HandlerFunc) {
	m.handlersMu.Lock()
	m.handlers[action] = handler
	m.handlersMu.Unlock()
}

// SendTo 发送消息给指定源的所有连接
func (m *Manager) SendTo(sourceID string, msg *Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for client := range m.clients[sourceID] {
		client.Send(msg)
	}
}

// SendToMany 发送消息给多个源
func (m *Manager) SendToMany(sourceIDs []string, msg *Message) {
	for _, sourceID := range sourceIDs {
		m.SendTo(sourceID, msg)
	}
}

// Broadcast 广播消息给所有连接
func (m *Manager) Broadcast(msg *Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, clients := range m.clients {
		for client := range clients {
			client.Send(msg)
		}
	}
}

// GetOnlineSourceIDs 获取所有在线源ID
func (m *Manager) GetOnlineSourceIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.clients))
	for id := range m.clients {
		ids = append(ids, id)
	}
	return ids
}

// GetOnlineCount 获取在线连接数
func (m *Manager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, clients := range m.clients {
		count += len(clients)
	}
	return count
}

// IsOnline 检查源是否在线
func (m *Manager) IsOnline(sourceID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, ok := m.clients[sourceID]
	return ok && len(clients) > 0
}

func (m *Manager) addClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[client.sourceID] == nil {
		m.clients[client.sourceID] = make(map[*Client]struct{})
	}
	m.clients[client.sourceID][client] = struct{}{}
}

func (m *Manager) removeClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, ok := m.clients[client.sourceID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(m.clients, client.sourceID)
		}
	}
}

func (m *Manager) handleMessage(client *Client, msg *Message) {
	m.handlersMu.RLock()
	handler, ok := m.handlers[msg.Action]
	m.handlersMu.RUnlock()

	if !ok {
		m.logger.Debug("no handler for action", zap.String("action", msg.Action))
		return
	}

	ctx := &Context{
		Client:  client,
		Message: msg,
		manager: m,
	}
	handler(ctx)
}
