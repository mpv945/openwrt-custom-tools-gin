package ws

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/model"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/json"
)

type Hub struct {
	Clients map[string]*Client
	Groups  map[string]map[string]*Client

	Broadcast chan []byte

	Register   chan *Client
	Unregister chan *Client

	mu sync.RWMutex
}

/*func NewHub() *Hub {

	return &Hub{
		Clients:    make(map[string]*Client),
		Groups:     make(map[string]map[string]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}*/

// DefaultHub 全局单例
var DefaultHub *Hub

// init 当包被导入时，这段代码会自动执行
//
//	在 main 中 不需要 再写 DefaultHub = NewHub() 或 DefaultHub.Run()
//
// init 初始化全局 Hub
func init() {
	DefaultHub = &Hub{
		Clients:    make(map[string]*Client),
		Groups:     make(map[string]map[string]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
	go DefaultHub.Run() // 启动 hub 事件循环
	go DefaultHub.StartCleaner()
	//go DefaultHub.Subscribe("ws_message")
}

func (h *Hub) Run() {

	for {

		select {

		case client := <-h.Register:

			h.mu.Lock()
			h.Clients[client.ID] = client

			if client.Group != "" {

				if h.Groups[client.Group] == nil {
					h.Groups[client.Group] = make(map[string]*Client)
				}

				h.Groups[client.Group][client.ID] = client
			}

			h.mu.Unlock()

		case client := <-h.Unregister:

			h.mu.Lock()

			delete(h.Clients, client.ID)

			if client.Group != "" {
				delete(h.Groups[client.Group], client.ID)
			}

			close(client.Send)

			h.mu.Unlock()

		case msg := <-h.Broadcast:

			h.mu.RLock()

			for _, client := range h.Clients {
				client.Send <- msg
			}

			h.mu.RUnlock()
		}
	}
}

// SendAll 广播
func SendAll(msg []byte) {
	// 然后分发给本地客户端
	DefaultHub.Broadcast <- msg // 或 SendToUser / SendToGroup
}

// SendTo 点对点消息
func SendTo(id string, msg []byte) {

	DefaultHub.mu.RLock()
	defer DefaultHub.mu.RUnlock()

	if c, ok := DefaultHub.Clients[id]; ok {
		c.Send <- msg
	}
}

// SendToGroup 群组消息
func SendToGroup(group string, msg []byte) {

	DefaultHub.mu.RLock()
	defer DefaultHub.mu.RUnlock()

	for _, c := range DefaultHub.Groups[group] {
		c.Send <- msg
	}
}

func Dispatch(msg model.Message) {

	DefaultHub.mu.RLock()
	defer DefaultHub.mu.RUnlock()

	toString, err := json.MarshalToString(msg)
	if err != nil {
		return
	}
	fmt.Println("redis 收到消息转发到本地: ", toString)
	switch msg.Type {

	case "broadcast":

		for _, c := range DefaultHub.Clients {
			c.Send <- []byte(msg.Data)
		}

	case "private":

		if c, ok := DefaultHub.Clients[msg.To]; ok {
			c.Send <- []byte(msg.Data)
		}

	case "group":

		for _, c := range DefaultHub.Groups[msg.To] {
			c.Send <- []byte(msg.Data)
		}
	}
}

// Hub 定期清理死连接
// go DefaultHub.StartCleaner()

func (h *Hub) StartCleaner() {

	ticker := time.NewTicker(60 * time.Second)

	for range ticker.C {

		h.mu.Lock()

		for id, client := range h.Clients {

			if time.Since(client.LastActive) > 120*time.Second {

				err := client.Conn.Close()
				if err != nil {
					return
				}

				delete(h.Clients, id)

			}
		}

		h.mu.Unlock()
	}
}

func (h *Hub) Shutdown() {
	log.Println("Hub shutdown started")

	h.mu.Lock()
	defer h.mu.Unlock()

	// 遍历所有客户端，发送下线通知(非阻塞)
	for _, client := range h.Clients {
		select {
		case client.Send <- []byte(`{"type":"offline","event":"server_shutdown","message":"Server shutting down"}`):
			// 消息已发送
		default:
			// client.Send 已满或不可写，跳过
		}
	}

	// 遍历客户端，非阻塞发送下线消息
	/*for _, client := range h.Clients {
		close(client.Send) // 关闭 channel，writePump 会退出
		err := client.Conn.Close()
		if err != nil {
			return
		} // 关闭 TCP 连接
	}*/

	// 清理 Clients map
	//h.Clients = make(map[string]*Client)

	log.Println("Hub shutdown finished")
}
