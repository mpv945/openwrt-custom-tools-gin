package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/config"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/db"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/router"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/scheduler"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/jwt"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/logger"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/mqtt"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/redis"
	"go.uber.org/zap"
)

func main() {
	// 生产环境： 默认启动 gin.Default() 时就是 debug 模式。test；适合单元测试，日志更少
	gin.SetMode(gin.ReleaseMode)
	//gin.SetMode(gin.TestMode)
	// 禁用控制台颜色
	// gin.DisableConsoleColor()

	// 初始化配置
	config.Init()

	// 初始化日志
	logger.Init()

	// 初始化token值
	jwt.InitJWT(config.Config.GetString("jwt.secret"))

	// 初始化数据库
	db.InitDB(
		config.Config.GetString("db.host"),
		config.Config.GetInt("db.port"),
		config.Config.GetString("db.user"),
		config.Config.GetString("db.password"),
		config.Config.GetString("db.dbname"),
		config.Config.GetString("db.driverName"),
	)

	// 初始化Redis
	redis.Init(config.Config.GetString("redis.addr"), config.Config.GetString("redis.password"))
	// 启动Redis Stream消费者
	go redis.StartConsumer()
	// 启动Pub/Sub 消费者
	go redis.Subscribe("order_channel", func(m redis.Message) {
		log.Println("Pub/Sub receive message:", m.Type, m.Data)
	})

	// 初始化MQTT
	mqtt.Init(
		config.Config.GetString("mqtt.broker"),
		config.Config.GetString("mqtt.clientId"),
		config.Config.GetString("mqtt.userName"),
		config.Config.GetString("mqtt.password"),
	)
	// 启动MQTT监听
	errMqtt := mqtt.Subscribe("device/+/status")
	if errMqtt != nil {
		log.Fatal(errMqtt)
	}

	// 启动HTTP(初始路由）
	r := router.InitRouter()

	// 启动定时任务
	scheduler.Start()

	// logger.Log.Info("Starting server...")
	logger.Log.Info("服务开始启动",
		zap.String("username", "Tom"),
		zap.Int("age", 20),
		zap.Bool("success", true),
	)

	// 生产环境，信任特定代理（如果有前置代理，值信任前置代理）
	err := r.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16"})
	if err != nil {
		return
	}

	port := config.Config.GetString("server.port")

	//r.Run(":" + port)
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := r.Run(":" + port); err != nil {
		logger.Log.Error(err.Error())
	}
}
