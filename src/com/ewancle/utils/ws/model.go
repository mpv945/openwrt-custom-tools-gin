package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID         string          // 用户ID
	Conn       *websocket.Conn // websocket连接
	Send       chan []byte     // 发送消息队列
	Group      string          // 所属群组
	Hub        *Hub            // 所属Hub
	LastActive time.Time
}
