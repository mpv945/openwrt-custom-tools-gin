package wsclient

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// 使用客服端
// 1 自动重连
//
//	for {
//		err := client.Connect()
//		if err != nil {
//			time.Sleep(5*time.Second)
//			continue
//		}
//	}
//
// 2 JWT header
// header := http.Header{}
// header.Add("Authorization", "Bearer "+token)
//
// websocket.DefaultDialer.Dial(url, header)
/*
	// HTTP 接口 -> 发送 websocket 消息
	r.GET("/send", func(c *gin.Context) {

		msg := c.Query("msg")

		wsClient.SendMessage(msg)

		c.JSON(200, gin.H{
			"msg": "sent",
		})
	})
*/

type Client struct {
	Url  string
	Conn *websocket.Conn
	Send chan []byte
	Done chan struct{}
}

// NewClient 创建客户端
func NewClient(addr string, path string) *Client {

	u := url.URL{
		Scheme: "ws",
		Host:   addr,
		Path:   path,
	}

	return &Client{
		Url:  u.String(),
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
}

// Connect 连接服务器
func (c *Client) Connect() error {

	conn, _, err := websocket.DefaultDialer.Dial(c.Url, nil)
	if err != nil {
		return err
	}

	c.Conn = conn

	log.Println("WebSocket connected:", c.Url)

	go c.readLoop()
	go c.writeLoop()

	return nil
}

// 接收消息
func (c *Client) readLoop() {

	defer func() {
		err := c.Conn.Close()
		if err != nil {
			return
		}
		close(c.Done)
	}()

	for {

		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}

		log.Println("rev:", string(message))
	}
}

// 发送消息
func (c *Client) writeLoop() {

	ticker := time.NewTicker(30 * time.Second)

	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	for {

		select {

		case msg := <-c.Send:

			err := c.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write error:", err)
				return
			}

		case <-ticker.C:

			// 心跳
			err := c.Conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("ping error:", err)
				return
			}
		}
	}
}

// SendMessage 发送消息接口
func (c *Client) SendMessage(msg string) {

	select {
	case c.Send <- []byte(msg):
	default:
		log.Println("send channel full")
	}
}
