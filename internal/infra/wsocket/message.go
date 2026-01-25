// Package wsocket 提供 WebSocket 连接管理和消息推送功能
package wsocket

import (
	"encoding/json"
)

// Message WebSocket 消息
type Message struct {
	Action    string          `json:"action"`               // 命令/动作标识
	RequestID string          `json:"request_id,omitempty"` // 请求ID（用于请求-响应模式）
	Data      json.RawMessage `json:"data,omitempty"`       // 业务数据
}

// NewMessage 创建消息
func NewMessage(action string, data any) *Message {
	msg := &Message{Action: action}
	if data != nil {
		if raw, err := json.Marshal(data); err == nil {
			msg.Data = raw
		}
	}
	return msg
}

// NewResponse 创建响应消息（带 RequestID）
func NewResponse(requestID, action string, data any) *Message {
	msg := NewMessage(action, data)
	msg.RequestID = requestID
	return msg
}

// Bytes 序列化为 JSON
func (m *Message) Bytes() []byte {
	b, _ := json.Marshal(m)
	return b
}

// ParseMessage 解析消息
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// UnmarshalData 解析 Data 到指定结构
func (m *Message) UnmarshalData(v any) error {
	if m.Data == nil {
		return nil
	}
	return json.Unmarshal(m.Data, v)
}
