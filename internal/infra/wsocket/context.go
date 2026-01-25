package wsocket

// Context 消息处理上下文
type Context struct {
	Client  *Client
	Message *Message
	manager *Manager
}

// SourceID 获取源标识
func (c *Context) SourceID() string {
	return c.Client.sourceID
}

// Reply 响应当前请求（保留 RequestID）
func (c *Context) Reply(data any) {
	msg := NewResponse(c.Message.RequestID, c.Message.Action, data)
	c.Client.Send(msg)
}

// Send 发送新消息给当前连接
func (c *Context) Send(action string, data any) {
	msg := NewMessage(action, data)
	c.Client.Send(msg)
}

// SendTo 发送消息给指定源
func (c *Context) SendTo(sourceID string, action string, data any) {
	msg := NewMessage(action, data)
	c.manager.SendTo(sourceID, msg)
}

// Broadcast 广播消息
func (c *Context) Broadcast(action string, data any) {
	msg := NewMessage(action, data)
	c.manager.Broadcast(msg)
}

// UnmarshalData 解析消息数据到指定结构
func (c *Context) UnmarshalData(v any) error {
	return c.Message.UnmarshalData(v)
}
