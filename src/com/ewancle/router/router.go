package router

/*import "github.com/gin-gonic/gin"

func InitRouter() *gin.Engine {

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}*/
import (
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/api"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {

	//r := gin.Default() // 每次请求接口会打印请求状态和耗时等信息
	// 如果想不显示，不要用 gin.Default()，改用 gin.New()。
	r := gin.New()
	// 只保留 panic 恢复
	r.Use(gin.Recovery())
	// 限制上传大小 32MB
	//r.MaxMultipartMemory = 32 << 20

	/*deviceAPI := new(device_api.DeviceAPI)

	device := r.Group("/device")
	{
		device.GET("/:id", deviceAPI.GetDevice)
		device.GET("/list", deviceAPI.ListDevices)
		device.POST("/create", deviceAPI.CreateDevice)
	}

	return r*/
	// 实例化每个业务 API
	deviceAPI := new(api.DeviceAPI)
	//userAPI := new(api.UserAPI)

	// 分组路由
	device := r.Group("/device")
	{
		/*device.GET("/:id", deviceAPI.GetDevice)
		device.POST("/create", deviceAPI.CreateDevice)*/
		device.GET("/:id", deviceAPI.GetDevice) // 访问路径 /device/123
		device.GET("/list", deviceAPI.ListDevices)
		device.POST("/create", deviceAPI.CreateDevice)
		device.GET("/export", deviceAPI.ExportExcel)
		device.POST("/upload", deviceAPI.UploadFile)     //curl -F "file=@test.zip" http://localhost:8080/upload
		device.POST("/download", deviceAPI.DownloadFile) //http://localhost:8080/download?file=test.zip
	}

	/*user := r.Group("/user")
	{
		user.GET("/:id", userAPI.GetUser)
		user.POST("/create", userAPI.CreateUser)
	}*/

	// 默认心跳接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}
