是的，WebSocket 本身是 **双向全双工通信协议**，客户端和服务器都可以随时发送和接收消息。但你可以通过 **逻辑约束** 实现“单向通信”，即只在一个方向发送消息。

// Create a WebSocket connection
const socket = new WebSocket("ws://localhost:8080");
// Event: Connection opened
socket.addEventListener("open", (event) => {
console.log("Connected to the server");
socket.send("Hello Server!"); // Send a message to the server
});
// Event: Message received from the server
socket.addEventListener("message", (event) => {
console.log("Message from server:", event.data);
});
// Event: Connection closed
socket.addEventListener("close", (event) => {
console.log("Disconnected from the server");
});
// Event: Error occurred
socket.addEventListener("error", (event) => {
console.error("WebSocket error:", event);
});

---

## 1️⃣ 客户端 → 服务端（上行单向）

* 客户端只发送消息给服务器，不处理服务器返回。
* 服务器可以接收消息，但不返回数据给客户端。

**示例（客户端）**

```javascript
const ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = () => {
  ws.send(JSON.stringify({ type: "message", content: "hello server" }));
};

// 不注册 onmessage
```

**服务端（Gin + Gorilla）**

```go
func serveWs(c *gin.Context) {
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    go func() {
        defer conn.Close()
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                break
            }
            log.Printf("Received: %s", string(msg))
        }
    }()
}
```

* 服务器没有调用 `WriteMessage` → 客户端收不到任何消息。

---

## 2️⃣ 服务端 → 客户端（下行单向）

* 服务器只主动向客户端发送消息。
* 客户端可以接收消息，但不发送任何数据。

**示例（客户端）**

```javascript
const ws = new WebSocket("ws://localhost:8080/ws");

ws.onmessage = (event) => {
  console.log("Server says:", event.data);
};

// 不调用 ws.send()
```

**服务端**

```go
func serveWs(c *gin.Context) {
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    go func() {
        defer conn.Close()
        for i := 0; i < 5; i++ {
            msg := fmt.Sprintf("Message %d from server", i)
            conn.WriteMessage(websocket.TextMessage, []byte(msg))
            time.Sleep(1 * time.Second)
        }
    }()
}
```

* 客户端只能接收消息。
* 如果客户端调用 `ReadMessage()`，可以保持阻塞等待，逻辑上就像单向。

---

## 3️⃣ 小结

* **WebSocket 协议**始终是双向的。
* **单向通信**是逻辑层面的约束，不需要协议改变。
* 你可以通过 **客户端或服务器不调用发送 API** 来实现单向消息流。

---

如果你希望，我可以帮你写一个 **生产级单向 WebSocket 模板**：

* 客户端到服务器单向（客服端 → 服务端）
* 服务器到客户端单向（广播系统消息）
* 保证安全关闭和心跳机制

这样比直接使用双向更安全、可控。

你希望我帮你写吗？
