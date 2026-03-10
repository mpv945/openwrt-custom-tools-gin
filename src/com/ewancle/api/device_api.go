package api

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mpv945/openwrt-custom-tools-gin/assets"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/service"
	base64util "github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/base64"
	commandutil "github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/command"
	cryptutil "github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/crypt"
	fileutil "github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/file"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/http_util"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/json"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/mqtt"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/redis"
	"github.com/xuri/excelize/v2"
)

type DeviceAPI struct{}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// GetDevice GET /device/:id
// Path 参数示例
func (d *DeviceAPI) GetDevice(c *gin.Context) {

	user := User{
		Name: "Tom",
		Age:  18,
	}
	// 序列化
	// 取第一个返回值，忽略第二个（通常是 error）;所以你必须显式用 _ 或变量捕获
	jsonStr, _ := json.MarshalToString(user)
	fmt.Println("User 序列化 JSON 字符串:", jsonStr)

	// 反序列化
	var u2 User
	err := json.UnmarshalFromString(jsonStr, &u2)
	if err != nil {
		// 处理错误
		fmt.Println("解析 JSON 失败:", err)
		return
	}
	/*if str, err := MarshalToString(obj); err == nil {
		fmt.Println(str)
	} else {
		fmt.Println("序列化失败:", err)
	}*/

	// HTTP GET 示例
	params := map[string]string{
		"id": "io.quarkus:quarkus-rest-client",
	}
	body, err := http_util.HttpGet(
		"https://stage.code.quarkus.io/api/extensions",
		params,
	)
	fmt.Println(string(body))

	// HTTP POST JSON 示例
	data := map[string]interface{}{
		"userId": 11,
		"id":     1,
		"title":  "测试",
		"body":   "sfdsfsdfsdfsfs",
	}
	resp, err := http_util.HttpPost(
		"https://jsonplaceholder.typicode.com/posts",
		data,
	)
	fmt.Println("保存数据", string(resp))

	// 编码
	encoded := base64util.EncodeString("hello go")
	fmt.Println("base64:", encoded)

	// 解码
	decoded, _ := base64util.DecodeString(encoded)
	fmt.Println("decode:", decoded)

	// 加密
	crypt, err := cryptutil.Sha256Crypt("cs33", "sfsdffdsfsd")
	if err != nil {
		return
	}
	fmt.Println("Sha256Crypt:", crypt)

	// 读取内联文件（解析excel文件内容）
	file, err := assets.FS.ReadFile("excel/report.xlsx")
	if err != nil {
		return
	}
	f, err := excelize.OpenReader(bytes.NewReader(file))
	if err != nil {
		panic(err)
	}
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		panic(err)
	}
	for _, row := range rows {

		for _, col := range row {
			fmt.Print(col, " ")
		}

		fmt.Println()
	}

	// 执行系统命令
	if runtime.GOOS == "windows" {
		resultComm := commandutil.ExecCommand("cmd", "/C", "dir")
		//resultComm := command_util.ExecCommand("powershell", "-Command", "Get-Process")

		fmt.Println(resultComm)
	} else {
		//resultComm := command_util.ExecCommand("ls","-l")
		resultComm := commandutil.ExecCommand("bash", "-c", "ls")
		fmt.Println(resultComm)
	}

	// 操作数据库
	userService := service.NewUserService()

	// 创建用户
	user10, err10 := userService.CreateUser(
		"Tom",
		"tom@example.com",
	)
	if err10 != nil {
		panic(err10)
	}
	//fmt.Println("create user:", user10.ID)
	fmt.Printf("ID=%d Name=%s Email=%s\n",
		user10.ID,
		user10.Name,
		user10.Email,
	)
	// 查询
	u, _ := userService.GetUser(user10.ID)
	fmt.Println("query user:", u.Name)

	// String
	err20 := redis.SetString("name", "redis-demo", time.Hour)
	if err20 != nil {
		return
	}

	v, _ := redis.GetString("name")
	log.Println("string:", v)

	// Hash
	err30 := redis.HSet("user:1", map[string]interface{}{
		"name": "tom",
		"age":  18,
	})
	if err30 != nil {
		return
	}

	user40, _ := redis.HGetAll("user:1")
	log.Println("hash:", user40)

	// List
	err60 := redis.LPush("task:list", "task1", "task2")
	if err60 != nil {
		return
	}

	list, _ := redis.LRange("task:list", 0, -1)
	log.Println("list:", list)

	// Set
	err70 := redis.SAdd("tags", "go", "redis", "mq")
	if err70 != nil {
		return
	}
	tags, _ := redis.SMembers("tags")
	log.Println("set:", tags)
	// 生产消息
	go func() {

		for i := 0; i < 10; i++ {

			message, err := redis.ProduceMessage(map[string]interface{}{
				"orderId": i,
				"price":   90000,
			})
			if err != nil {
				return
			}
			fmt.Println("发送 id = ", message)

			time.Sleep(time.Second)
		}

	}()

	// Pub/Sub消息发送
	// 发送消息
	for i := 0; i < 5; i++ {

		msg := redis.Message{
			Type: "order_create",
			Data: map[string]interface{}{
				"orderId": i,
				"price":   10000,
			},
		}

		err := redis.Publish("order_channel", msg)
		if err != nil {
			log.Println(err)
		}

		time.Sleep(2 * time.Second)
	}

	// ---------------- 布隆过滤器（Bloom Filter） → 防止重复 / 缓存穿透 ----------------
	bloom := redis.NewBloomFilter(redis.Client, "bloom:user", 1000000)
	err91 := bloom.Add("user1001")
	if err91 != nil {
		return
	}
	exist, _ := bloom.Exists("user1001")
	fmt.Println("bloom exists:", exist)
	// ---------------- UV（Unique Visitor）统计 → 使用 Redis HyperLogLog ----------------
	uv := redis.NewUVService(redis.Client)
	err92 := uv.AddUV("uv:20260310", "user1")
	if err92 != nil {
		return
	}
	_ = uv.AddUV("uv:20260310", "user2")
	_ = uv.AddUV("uv:20260310", "user1")
	countUV, _ := uv.CountUV("uv:20260310")
	fmt.Println("UV:", countUV)

	// ---------------- PV（Page View）统计 → 使用 Redis 计数器 ----------------

	pv := redis.NewPVService(redis.Client)
	incrPV, err94 := pv.IncrPV("pv:home")
	if err94 != nil {
		return
	}
	fmt.Println(incrPV)

	_, _ = pv.IncrPV("pv:home")

	pvCount, _ := pv.GetPV("pv:home")
	fmt.Println("PV:", pvCount)

	// ---------------- 限流 ----------------
	limiter := redis.NewSlidingWindowLimiter(
		redis.Client,
		5,
		time.Second*10,
	)
	for i := 0; i < 10; i++ {
		allowed, _ := limiter.Allow("rate:user:1001")
		fmt.Println("request", i, "allowed:", allowed)
		time.Sleep(time.Second)
	}

	// ---------------- 排行榜 ----------------

	board := redis.NewLeaderboard(redis.Client, "game:rank")

	err98 := board.AddScore("user1", 100)
	if err98 != nil {
		return
	}
	err99 := board.AddScore("user2", 200)
	if err99 != nil {
		return
	}
	err100 := board.AddScore("user3", 150)
	if err100 != nil {
		return
	}

	// 只保留前10000名(定期清理）
	/*err101 := board.Trim(100)
	if err101 != nil {
		return
	}*/

	top, _ := board.TopN(3)
	fmt.Println("Top players:")
	for _, v := range top {
		fmt.Println(v.Member, v.Score)
	}
	rank, _ := board.Rank("user1")
	fmt.Println("user1 rank:", rank)

	//id := uuid.New()
	/*id := uuid.Must(uuid.NewV7())
	fmt.Println(id.String())
	ack := map[string]any{
		"msgId":  id,
		"status": "ok",
	}
	byteArr, _ := json.Marshal(ack)*/
	err103 := mqtt.Publish(
		"device/1001/status",
		[]byte(`{"temp":22}`),
		//byteArr,
	)
	if err103 != nil {
		return
	}

	// 分布式锁
	// 创建分布式锁
	lock := redis.NewDistributedLock("lock:order:1001")
	// 获取锁
	if err := lock.Lock(); err != nil {
		log.Println("获取锁失败:", err)
		return
	}
	log.Println("获取锁成功")
	defer func() {
		ok, err := lock.Unlock()
		if err != nil || !ok {
			log.Println("释放锁失败")
		} else {
			log.Println("锁已释放")
		}

	}()
	// 模拟业务处理
	log.Println("执行业务逻辑")
	time.Sleep(5 * time.Second)

	log.Println("业务执行完成")

	// Path 参数示例: GET /device/:id
	id := c.Param("id")
	// Query 参数示例: GET /device/list?page=1&size=10
	//page := c.DefaultQuery("page", "1")
	//size := c.DefaultQuery("size", "10")
	/*c.JSON(http.StatusOK, gin.H{
		"page": page,
		"size": size,
		"list": []string{"device1", "device2"},
	})*/

	// POST /device/create
	// JSON参数示例
	/*type CreateDeviceRequest struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
		var req CreateDeviceRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "device created",
		"name": req.Name,
		"type": req.Type,
	})
	*/

	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"name":      "device-" + id,
		"client_ip": c.ClientIP(),
	})
}

// ListDevices GET /device/list?page=1&size=10
// Query 参数示例
func (d *DeviceAPI) ListDevices(c *gin.Context) {

	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	c.JSON(http.StatusOK, gin.H{
		"page": page,
		"size": size,
		"list": []string{"device1", "device2"},
	})
}

// CreateDeviceRequest POST /device/create
// JSON参数示例
type CreateDeviceRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (d *DeviceAPI) CreateDevice(c *gin.Context) {

	var req CreateDeviceRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "device created",
		"name": req.Name,
		"type": req.Type,
	})
}

func (d *DeviceAPI) ExportExcel(c *gin.Context) {

	f := excelize.NewFile()
	sw, _ := f.NewStreamWriter("Sheet1")
	header := []interface{}{"ID", "Name", "Email"}
	err := sw.SetRow("A1", header)
	if err != nil {
		return
	}

	for i := 1; i <= 10000; i++ {
		row := []interface{}{
			i,
			fmt.Sprintf("user_%d", i),
			fmt.Sprintf("user_%d@test.com", i),
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+1)

		err := sw.SetRow(cell, row)
		if err != nil {
			return
		}
	}

	err2 := sw.Flush()
	if err2 != nil {
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=report.xlsx")

	c.Header("Content-Transfer-Encoding", "binary")

	err3 := f.Write(c.Writer)
	if err3 != nil {
		c.String(http.StatusInternalServerError, err3.Error())
	}
}

func (d *DeviceAPI) UploadFile(c *gin.Context) {

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	filename := filepath.Base(header.Filename)

	path, err := fileutil.SaveFile(file, filename)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"file": path,
	})
}

func (d *DeviceAPI) DownloadFile(c *gin.Context) {

	name := c.Query("file")

	path := filepath.Join("./uploads", filepath.Base(name))

	f, err := os.Open(path)
	if err != nil {
		c.String(404, "file not found")
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	stat, err := f.Stat()
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+stat.Name())
	c.Header("Content-Type", "application/octet-stream")

	http.ServeContent(
		c.Writer,
		c.Request,
		stat.Name(),
		stat.ModTime(),
		f,
	)
}
