package sse

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/json"

	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/redis"
)

// 连接管理（manager.go）

/*
如果你用 Nginx，必须关闭缓冲：
location /sse {

    proxy_pass http://backend;

    proxy_http_version 1.1;
    proxy_set_header Connection "";

    proxy_buffering off;
    proxy_cache off;

    chunked_transfer_encoding on;
}

TCP优化
net.core.somaxconn = 65535
net.ipv4.tcp_tw_reuse = 1
*/

type Manager struct {
	clients map[string]*Client

	users map[string]map[string]*Client

	topics map[string]map[string]*Client

	lock sync.RWMutex
}

/*var GlobalManager *Manager*/

var (
	globalManager *Manager
	once          sync.Once
)

func GetManager() *Manager {

	once.Do(func() {
		globalManager = &Manager{
			clients: map[string]*Client{},
			users:   map[string]map[string]*Client{},
			topics:  map[string]map[string]*Client{},
		}
	})

	return globalManager
}

func (m *Manager) Add(c *Client) {

	m.lock.Lock()

	defer m.lock.Unlock()

	m.clients[c.ID] = c

	if m.users[c.UserID] == nil {
		m.users[c.UserID] = map[string]*Client{}
	}

	m.users[c.UserID][c.ID] = c
}

func (m *Manager) Remove(id string) {

	m.lock.Lock()
	defer m.lock.Unlock()

	c := m.clients[id]

	if c == nil {
		return
	}

	delete(m.clients, id)

	delete(m.users[c.UserID], id)

	for t := range c.Topics {

		delete(m.topics[t], id)

		if len(m.topics[t]) == 0 {
			delete(m.topics, t)
		}
	}

	close(c.Ch)
}

func (m *Manager) Subscribe(c *Client, topic string) {

	m.lock.Lock()

	defer m.lock.Unlock()

	if m.topics[topic] == nil {
		m.topics[topic] = map[string]*Client{}
	}

	m.topics[topic][c.ID] = c

	c.Topics[topic] = true
}

// 用户推送

func (m *Manager) SendUser(user string, e Event) {

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, c := range m.users[user] {

		select {
		case c.Ch <- e:
		default:
		}
	}
}

// Topic 推送

func (m *Manager) SendTopic(topic string, e Event) {

	m.lock.RLock()
	defer m.lock.RUnlock()
	size := len(m.topics)
	fmt.Println("当前有: ", size, " 查询topic: ", topic)
	for _, c := range m.topics[topic] {

		select {
		case c.Ch <- e:
		default:
		}
	}
}

// 广播

func (m *Manager) Broadcast(e Event) {

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, c := range m.clients {

		select {
		case c.Ch <- e:
		default:
		}
	}
}

// Redis PubSub 分布式广播

func Publish(channel string, e Event) error {
	data, _ := json.Marshal(e)
	return redis.Client.Publish(redis.Ctx, channel, data).Err()
}

// 节点收到消息：

func (m *Manager) SseSubscribe(channel string) {

	PubNub := redis.Client.Subscribe(redis.Ctx, channel)

	ch := PubNub.Channel()

	go func() {

		for msg := range ch {

			var e Event

			err := json.Unmarshal([]byte(msg.Payload), &e)
			if err != nil {
				return
			}

			toString, err := json.MarshalToString(e)
			if err != nil {
				return
			}
			fmt.Println("sse订阅收到内容： " + toString)
			if e.UserID != "" {

				m.SendUser(e.UserID, e)

			} else if e.Topic != "" {

				m.SendTopic(e.Topic, e)

			} else {

				m.Broadcast(e)
			}
		}
	}()
}

func Handler(manager *Manager) gin.HandlerFunc {

	return func(c *gin.Context) {

		w := c.Writer
		r := c.Request

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			c.String(http.StatusInternalServerError, "stream unsupported")
			return
		}

		flusher.Flush()

		id := uuid.New().String()
		//user := uuid.New().String()
		//user := r.URL.Query().Get("userId")
		user := "1001"
		// 检测token
		/* token := r.URL.Query().Get("token")
		claims, err := jwt.ParseToken(token)
		if err != nil {
			http.Error(w, "unauthorized", 401)
			return
		}
		client := NewClient(id, claims.UserID, w)*/

		client := NewClient(id, user, w)

		manager.Add(client)

		log.Println("SSE connected:", id)

		defer manager.Remove(id)

		// 自动订阅 topic
		fmt.Println("URL:", r.URL.String())
		fmt.Println("RawQuery:", r.URL.RawQuery)
		fmt.Println("topic:", r.URL.Query().Get("topic"))
		topics := c.Query("topic")

		fmt.Println("获取到topic: ", topics)
		if topics != "" {

			for _, t := range strings.Split(topics, ",") {

				manager.Subscribe(client, strings.TrimSpace(t))
			}
		}

		// 返回 clientId
		_, err := fmt.Fprintf(w, "event: init\ndata: {\"clientId\":\"%s\"}\n\n", id)
		if err != nil {
			return
		}
		flusher.Flush()

		go client.Loop(r)

		<-client.Done
	}
}

func SubscribeAPI(manager *Manager) gin.HandlerFunc {

	return func(c *gin.Context) {
		//w := c.Writer
		r := c.Request
		id := r.URL.Query().Get("clientId")

		topic := r.URL.Query().Get("topic")

		manager.lock.RLock()

		cid := manager.clients[id]

		manager.lock.RUnlock()

		if c != nil {
			manager.Subscribe(cid, topic)
		}
	}
}
