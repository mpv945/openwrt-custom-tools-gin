你定义的泛型结构体：

```go
type Message[T any] struct {
	MsgId string `json:"msgId"`
	Data  T      `json:"data"`
}
```

这是 **Go 1.18+ 泛型结构体**，非常适合 **MQTT / Kafka / HTTP 消息模型**。下面给你完整的使用方式。

---

# 一、定义具体业务结构体

例如订单消息：

```go
type Order struct {
	OrderId string `json:"orderId"`
	Amount  int    `json:"amount"`
}
```

---

# 二、创建 Message

使用时要 **指定泛型类型**：

```go
msg := Message[Order]{
	MsgId: "123456",
	Data: Order{
		OrderId: "A1001",
		Amount:  200,
	},
}
```

---

# 三、序列化 JSON

```go
bytes, err := json.Marshal(msg)
if err != nil {
	log.Fatal(err)
}

fmt.Println(string(bytes))
```

输出：

```json
{
  "msgId": "123456",
  "data": {
    "orderId": "A1001",
    "amount": 200
  }
}
```

---

# 四、反序列化 JSON

```go
var msg Message[Order]

err := json.Unmarshal(payload, &msg)
if err != nil {
	log.Fatal(err)
}

fmt.Println(msg.MsgId)
fmt.Println(msg.Data.OrderId)
```

---

# 五、MQTT 消费场景（推荐写法）

例如 MQTT handler：

```go
func handler(client mqtt.Client, msg mqtt.Message) {

	var m Message[Order]

	err := json.Unmarshal(msg.Payload(), &m)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("msgId:", m.MsgId)
	fmt.Println("orderId:", m.Data.OrderId)

}
```

---

# 六、不同消息类型复用

泛型最大优势是 **不同 Data 类型复用 Message**。

例如设备数据：

```go
type DeviceData struct {
	DeviceId string  `json:"deviceId"`
	Temp     float64 `json:"temp"`
}
```

使用：

```go
msg := Message[DeviceData]{
	MsgId: "abc123",
	Data: DeviceData{
		DeviceId: "10001",
		Temp: 23.5,
	},
}
```

---

# 七、如果 Data 不确定

可以使用：

```go
Message[map[string]any]
```

例如：

```go
var msg Message[map[string]any]

json.Unmarshal(payload, &msg)

fmt.Println(msg.Data["orderId"])
```

---

# 八、MQTT ACK 场景（非常推荐）

MQTT 可靠消息通常这样设计：

### 发送消息

```go
type Order struct {
	OrderId string `json:"orderId"`
}

msg := Message[Order]{
	MsgId: uuid.NewString(),
	Data: Order{
		OrderId: "10001",
	},
}
```

发布：

```go
payload, _ := json.Marshal(msg)

client.Publish("order/create", 1, false, payload)
```

---

### 消费消息

```go
var m Message[Order]

json.Unmarshal(msg.Payload(), &m)

process(m.Data)
```

---

### ACK

```go
type Ack struct {
	MsgId  string `json:"msgId"`
	Status string `json:"status"`
}
```

---

# 九、推荐的完整消息模型

生产项目我通常这样设计：

```go
type Message[T any] struct {
	MsgId     string `json:"msgId"`
	Timestamp int64  `json:"timestamp"`
	Data      T      `json:"data"`
}
```

使用：

```go
msg := Message[Order]{
	MsgId:     uuid.NewString(),
	Timestamp: time.Now().Unix(),
	Data: Order{
		OrderId: "10001",
	},
}
```

---

# 十、总结

使用步骤：

1️⃣ 定义泛型结构体

```go
type Message[T any]
```

2️⃣ 定义业务结构

```go
type Order struct{}
```

3️⃣ 使用时指定类型

```go
Message[Order]
```

---

推荐使用场景：

| 场景       | 推荐    |
| -------- | ----- |
| MQTT消息   | ⭐⭐⭐⭐⭐ |
| Kafka消息  | ⭐⭐⭐⭐⭐ |
| HTTP API | ⭐⭐⭐⭐  |
| RPC      | ⭐⭐⭐⭐  |

---

如果你在做 **MQTT + ACK 可靠消息系统**，我可以给你一套 **完整生产级消息模型设计**（非常适合 IoT / 实时系统）：

包括：

* `Message[T]`
* `AckMessage`
* `RetryMessage`
* `ErrorMessage`
* `TraceMessage`

这样你的 **MQTT 系统可靠性会接近 Kafka**。
