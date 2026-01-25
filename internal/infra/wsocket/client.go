package wsocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入超时
	writeWait = 10 * time.Second
	// 读取 pong 超时
	pongWait = 60 * time.Second
	// ping 间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 512 * 1024 // 512KB
	// 发送缓冲区大小
	sendBufferSize = 256
)

// Client 单个 WebSocket 连接
type Client struct {
	manager  *Manager
	conn     *websocket.Conn
	sourceID string      // 源标识（如用户ID）
	send     chan []byte // 发送缓冲

	closeOnce sync.Once
	closed    chan struct{}
}

// newClient 创建新连接
func newClient(manager *Manager, conn *websocket.Conn, sourceID string) *Client {
	return &Client{
		manager:  manager,
		conn:     conn,
		sourceID: sourceID,
		send:     make(chan []byte, sendBufferSize),
		closed:   make(chan struct{}),
	}
}

// SourceID 获取源标识
func (c *Client) SourceID() string {
	return c.sourceID
}

// Send 发送消息
func (c *Client) Send(msg *Message) {
	select {
	case c.send <- msg.Bytes():
	case <-c.closed:
	default:
		// 缓冲区满，丢弃消息
	}
}

// Close 关闭连接
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		_ = c.conn.Close()
		c.manager.unregister <- c
	})
}

// readPump 读取消息循环
func (c *Client) readPump() {
	defer c.Close()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		msg, err := ParseMessage(data)
		if err != nil {
			continue
		}

		c.manager.handleMessage(c, msg)
	}
}

// writePump 写入消息循环
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case data := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.closed:
			return
		}
	}
}
