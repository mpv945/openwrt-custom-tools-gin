package sse

// 事件结构（event.go）

type Event struct {
	ID     string `json:"id"`
	Event  string `json:"event"`
	Data   string `json:"data"`
	UserID string `json:"userId,omitempty"`
	Topic  string `json:"topic,omitempty"`
}
