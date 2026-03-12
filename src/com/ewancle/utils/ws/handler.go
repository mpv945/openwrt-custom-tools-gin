package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/model"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/json"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/redis"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/*
| 参数            | 建议值 |
| ------------- | --- |
| ping interval | 30s |
| pong wait     | 60s |
| write timeout | 10s |

server 每30秒 ping
client 返回 pong
如果60秒没有 pong
连接关闭
*/
// 定义心跳常量 (pingPeriod < pongWait)
const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 5120
)

func ServeWs() gin.HandlerFunc {

	// 开启消息压缩,可降低带宽。
	// 代理设置: proxy_read_timeout 3600;
	upgrader.EnableCompression = true

	return func(c *gin.Context) {

		id := c.Query("id")
		group := c.Query("group")

		// 验证token
		/*token := c.Query("token")
		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			return
		}
		fmt.Println("claims:", claims.UserID, claims.Roles)*/

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &Client{
			ID:         id,
			Group:      group,
			Conn:       conn,
			Send:       make(chan []byte, 256),
			Hub:        DefaultHub,
			LastActive: time.Now(),
		}

		DefaultHub.Register <- client

		go writePump(client)
		go readPump(DefaultHub, client)
	}
}

/*
三、readPump 和 writePump 配合工作
	readPump：不断从 WebSocket 连接中读取客户端消息，当接收到消息时，会 处理 或 广播，然后将处理结果放入 Send channel，等待 writePump 发送给客户端。
	writePump：从 Send channel 获取消息并通过 WebSocket 发送给客户端，同时还要定期发送 Ping 消息来维持连接。

总结：
	readPump 是 读取 消息，writePump 是 发送 消息。
	它们的作用配合可以 实时双向通信，保证 WebSocket 连接是双向活跃的：客户端可以发送数据，服务器可以响应数据。
*/

// 读取消息
/*
一、readPump 的作用（读取 WebSocket 消息）
	职责：readPump 负责 从客户端读取消息，即接收客户端发送到 WebSocket 的数据。

工作原理：
	它从 WebSocket 连接中读取消息，通常是文本或二进制数据。
	收到的消息可以是 普通消息、Ping/Pong 心跳、关闭连接请求等。
	对于每个收到的消息，可以进行 解析，然后将其发送到适当的处理逻辑，或者广播给其他客户端。

作用总结：
	用于 接收 客户端发送的消息。
	对 消息进行解析 和 处理。
	常见用途包括 广播消息、私聊消息、群聊消息等。
*/
func readPump(hub *Hub, client *Client) {

	defer func() {
		hub.Unregister <- client
		err := client.Conn.Close()
		if err != nil {
			return
		}
	}()

	client.Conn.SetReadLimit(maxMessageSize)

	err := client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}

	client.Conn.SetPongHandler(func(string) error {
		// 收到 Pong 时 刷新连接超时。
		err := client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		client.LastActive = time.Now()
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			// 如果是连接关闭或错误
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("关闭了:Unexpected close from client %s: %v", client.ID, err)
			}
			// 打印事件日志
			log.Printf("Client %s disconnected due to error: %v", client.ID, err)
			break
		}

		// 更新活跃时间
		client.LastActive = time.Now()

		var msg model.Message

		err = json.Unmarshal(message, &msg)
		if err != nil {
			continue
		}
		msg.From = client.ID

		/*toString, err := json.MarshalToString(msg)
		if err != nil {
			return
		}
		fmt.Println("开始读取ws消息: ", toString)*/
		// 收到消息 -> 广播
		//hub.Broadcast <- message
		// 关键：发布到 Redis
		err11 := Publish("ws_message", msg)
		if err11 != nil {
			return
		}
	}
}

// 写入消息
/*
二、writePump 的作用（向 WebSocket 发送消息）
	职责：writePump 负责 向客户端写入消息，即将数据从服务器推送到客户端。

工作原理：
	它从 WebSocket 客户端的 发送缓冲区（Send channel）中读取消息，并将消息通过 WebSocket 连接写入客户端。
	writePump 也负责 Ping/Pong 心跳机制，定期向客户端发送 Ping 消息，以保持连接的活跃状态。

作用总结：
	用于 发送 消息给客户端，通常是 响应客户端请求 或 广播服务器端消息。
	负责 Ping/Pong 心跳，定期与客户端保持连接，避免因长时间未发送数据而被关闭。
*/
func writePump(client *Client) {

	ticker := time.NewTicker(pingPeriod)

	defer func() {

		ticker.Stop()
		err := client.Conn.Close()
		if err != nil {
			return
		}

	}()

	for {

		select {

		// 从发送通道获取消息
		case msg, ok := <-client.Send:

			// 设置写操作的超时时间
			err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}

			if !ok {
				// 向客户端发送 Ping 消息
				err := client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("发送测试Ping消息 Error writing message to client %s: %v", client.ID, err)
					return
				}
				return
			}

			// 向客户端写入消息
			err11 := client.Conn.WriteMessage(websocket.TextMessage, msg)

			if err11 != nil {
				log.Printf("发送消息失败: 网络断开、WiFi切换 Error writing message to client %s: %v", client.ID, err)
			}
		// 每隔一段时间发送 Ping 消息以维持连接
		case <-ticker.C:

			err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}

			// 向客户端发送 Ping 消息
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping 失败:网络断开、WiFi切换 Error writing message to client %s: %v", client.ID, err)
				return
			}

		}
	}
}

// 发送ws消息: 当 某个节点收到 WebSocket 消息，就调用

func Publish(channel string, msg model.Message) error {
	data, _ := json.Marshal(msg)
	return redis.Client.Publish(redis.Ctx, channel, data).Err()
}
func Subscribe(channel string) {

	sub := redis.Client.Subscribe(redis.Ctx, channel)
	go func() {
		ch := sub.Channel()
		for msg := range ch {
			var m model.Message

			err := json.Unmarshal([]byte(msg.Payload), &m)
			if err != nil {
				log.Println("redis msg parse error:", err)
				continue
			}
			// 关键：分发给本节点 websocket
			fmt.Println("ws redis收到信息:")
			Dispatch(m)
		}
	}()
}

func HandleShutdown() {
	// 发送关闭通知给所有连接的客户端
	/*for _, client := range DefaultHub.Clients {
		client.Send <- []byte(`{"event": "server_shutdown", "data": "Server is shutting down"}`)
	}*/
	// 这里假设你有 Redis 客户端 RedisClient
	//message := []byte("Server shutting down.")
	/*msg := model.Message{
		Type: "broadcast",
		//From:      "user1",
		//To:        "user2",
		Data: "Server shutting down.",
		//Content:   "hello",
		//Timestamp: time.Now().Unix(),
	}
	err := Publish("ws_message", msg)
	if err != nil {
		fmt.Println(err)
	}*/
	// 停止Hub等清理操作
	DefaultHub.Shutdown()
}
