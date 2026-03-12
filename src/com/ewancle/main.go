package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/config"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/db"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/router"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/scheduler"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/jwt"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/logger"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/mqtt"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/redis"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/sse"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/ws"
	"go.uber.org/zap"
)

func main() {

	// 设置最大 CPU 核心数
	runtime.GOMAXPROCS(runtime.NumCPU())

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
	r := router.InitHttpRouter()

	// Nginx 必须配置（SSE/WS）
	/*
						location /sse {

						    proxy_http_version 1.1;

						    proxy_set_header Connection "";

						    proxy_buffering off;

						    proxy_cache off;

						    proxy_read_timeout 1h;
						}

					location /ws {

					    proxy_http_version 1.1;

					    proxy_set_header Upgrade $http_upgrade;
					    proxy_set_header Connection "upgrade";

					    proxy_read_timeout 1h;
					}
		内核配置
		ulimit -n 200000
		sysctl 配置
		net.core.somaxconn = 65535
		net.ipv4.tcp_tw_reuse = 1
		net.ipv4.ip_local_port_range = 1024 65535
	*/

	// 启动WS和SSE(初始路由）
	rs := router.InitStreamRouter()

	// ws 总线 循环事件监听启动
	// 启动ws的redis消息
	go ws.Subscribe("ws_message")
	/*hub := ws.NewHub()
	go hub.Run()*/
	// 启动ws服务(访问路径/ws
	rs.GET("/ws", ws.ServeWs())

	// sse服务
	sseManager := sse.GetManager()
	// 启动sse分布式监听
	sseManager.SseSubscribe("sse_message")

	//http.HandleFunc("/sse", sse.Handler(sseManager))
	rs.GET("/sse", sse.Handler(sseManager))

	//http.HandleFunc("/subscribe", sse.SubscribeAPI(sseManager))
	rs.POST("/subscribe", sse.SubscribeAPI(sseManager))
	//r.GET("/sse", sseHandler)
	//r.GET("/send", sendMessage)

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

	errs := rs.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16"})
	if errs != nil {
		return
	}
	socketPort := config.Config.GetString("socket.port")
	// 普通启动
	//r.Run(":" + port)
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	/*if err := r.Run(":" + port); err != nil {
		logger.Log.Error(err.Error())
	}*/

	// 带优雅关闭的
	// 创建自定义 HTTP Server
	srv := &http.Server{
		Addr:    ":" + port, // 监听端口
		Handler: r,          // Gin 路由器
		//ReadTimeout: 10 * time.Second,
		// WriteTimeout: 10 * time.Second,
		//WriteTimeout: 0, // 关键 sse 必须设置成 0
		//IdleTimeout:  60 * time.Second,
		//| 配置                | 建议   | 原因          |
		//| ----------------- | ---- | ----------- |
		//| ReadTimeout       | 30s  | 防止慢客户端攻击    |
		//| WriteTimeout      | 70s  | 支持60s接口     |
		//| ReadHeaderTimeout | 10s  | 防止Slowloris |
		//| IdleTimeout       | 120s | KeepAlive稳定 |
		// 请求读取
		ReadTimeout: 30 * time.Second,

		// 响应写入（关键）
		WriteTimeout: 70 * time.Second,

		// header读取
		ReadHeaderTimeout: 10 * time.Second,

		// keepalive
		IdleTimeout: 120 * time.Second,

		// | 表达式       | 实际大小 |
		//| --------- | ---- |
		//| `1 << 10` | 1 KB |
		//| `1 << 20` | 1 MB |
		//| `1 << 30` | 1 GB |
		MaxHeaderBytes: 1 << 20,
	}

	socketSrv := &http.Server{
		Addr:              ":" + socketPort, // 监听端口
		Handler:           rs,               // Gin 路由器
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		BaseContext: func(net.Listener) context.Context {
			return context.Background()
		},
	}

	// 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Http ListenAndServe failed: %v", err)
		}
	}()
	go func() {
		if err := socketSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Socket ListenAndServe failed: %v", err)
		}
	}()

	// 延迟一点时间等待服务启动
	/*url := "http://127.0.0.1:" + port + "/ws/test"
	go func() {
		time.Sleep(1000 * time.Millisecond)

		if err := browser.Open(url); err != nil {
			log.Printf("浏览器打开失败: %v", err)
		}
	}()*/

	//| 事件          | 信号      |
	//| ----------- | ------- |
	//| Ctrl+C      | SIGINT  |
	//| Linux kill  | SIGTERM |
	//| Docker stop | SIGTERM |
	//| K8s Pod 删除  | SIGTERM |
	//| kill -9     | ❌无法捕获   |
	// 等待中断信号来优雅关闭服务
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // 阻塞直到接收到信号

	log.Println("Shutdown signal received")
	// 关闭资源
	ws.HandleShutdown()

	// 设置超时关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
	if err := socketSrv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}

	log.Println("Server exiting")
}
