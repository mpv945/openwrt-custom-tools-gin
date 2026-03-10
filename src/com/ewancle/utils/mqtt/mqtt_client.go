package mqtt

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var Client mqtt.Client

/*type MQTTClient struct {
	Client mqtt.Client
}*/

//func Init(broker string, clientID string, username string, password string) *MQTTClient {

func Init(broker string, clientID string, username string, password string) {

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)

	// 自动重连
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	// 心跳
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	// 断线消息缓存
	opts.SetCleanSession(false)

	// 连接成功回调
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT Connected")
	}

	// 连接丢失
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Println("MQTT Connection lost:", err)
	}

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {

		fmt.Println("Topic:", msg.Topic())
		fmt.Println("Message:", string(msg.Payload()))

	})

	Client = mqtt.NewClient(opts)

	/*token := Client.Connect()
	token.Wait()

	if token.Error() != nil {
		panic(token.Error())
	}*/
	token := Client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	/*return &MQTTClient{
		Client: Client,
	}*/
}

func Publish(topic string, payload []byte) error {

	token := Client.Publish(
		topic,
		1,     // QoS
		false, // retain
		payload,
	)

	if token.WaitTimeout(5 * time.Second) {
		return token.Error()
	}
	return fmt.Errorf("publish timeout")
}

func Subscribe(topic string) error {

	handler := func(client mqtt.Client, msg mqtt.Message) {

		go func() {

			//payload := string(msg.Payload())

			/*var data struct {
				MsgId string `json:"msgId"`
				Data  any    `json:"data"`
			}

			json.Unmarshal(msg.Payload(), &data)*/

			//log.Printf("topic=%s payload=%s\n", msg.Topic(), payload)

			// TODO 业务处理
			//processMessage(payload)
			// 业务处理
			err := process(msg.Payload())

			if err == nil {

				// 手动 ACK
				msg.Ack()

			} else {

				// 不 ACK
				// Broker 会重新投递
			}

		}()

	}
	token := Client.Subscribe(topic, 1, handler)
	token.Wait()
	return token.Error()
}

func process(data []byte) error {

	fmt.Println("processing:", string(data))

	return nil
}

func processMessage(msg string) {
	fmt.Println("MQTT Receive:", msg)
}
