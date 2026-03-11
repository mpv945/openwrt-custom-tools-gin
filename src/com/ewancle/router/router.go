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
	"html/template"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/mpv945/openwrt-custom-tools-gin/assets"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/api"
	auth "github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {

	//r := gin.Default() // 每次请求接口会打印请求状态和耗时等信息
	// 如果想不显示，不要用 gin.Default()，改用 gin.New()。
	r := gin.New()

	// 跨域处理
	r.Use(cors.New(cors.Config{

		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},

		ExposeHeaders: []string{
			"Content-Length",
		},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}))

	// 只保留 panic 恢复
	r.Use(gin.Recovery())
	// 限制上传大小 32MB
	//r.MaxMultipartMemory = 32 << 20

	// 认证路由组
	//authorized := r.Group("/", AuthRequired())

	// router.Static("/static", "/var/www")
	//r.Static("/assets", "./assets")

	//tl := template.Must(template.New("").ParseFS(assets.HTML, "html/*.tmpl", "html/**/*.tmpl"))
	tl := template.Must(template.New("").ParseFS(assets.HTML, "html/*.tmpl"))
	r.SetHTMLTemplate(tl)

	// html 模板使用需要引用就是使用 /public/static/**/*.* 文件
	r.StaticFS("/public", http.FS(assets.STATIC))

	r.GET("/login", func(c *gin.Context) {
		//c.HTML(200, "login.html", nil)
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"title": "Schedule Manager",
		})
	})
	r.POST("/api/login", func(c *gin.Context) {

		user := c.PostForm("username")
		pass := c.PostForm("password")

		if user == "admin" && pass == "123456" {

			c.JSON(200, gin.H{
				"code": 0,
			})

			return
		}

		c.JSON(200, gin.H{
			"code":    1,
			"message": "Invalid username or password",
		})
	})

	//  登录前端伪代码
	//fetch("http://localhost:8080/login", { method: "POST" })
	//.then(res => res.json())
	//.then(data => {
	//    const token = data.token
	//
	//    // 获取 profile
	//    fetch("http://localhost:8080/profile", {
	//        headers: { "Authorization": "Bearer " + token }
	//    }).then(res => res.json()).then(console.log)
	//
	//    // 获取 admin
	//    fetch("http://localhost:8080/admin", {
	//        headers: { "Authorization": "Bearer " + token }
	//    }).then(res => res.json()).then(console.log)
	//})

	// 普通接口，所有登录用户都可访问
	r.GET("/profile", auth.JWTAuth(), func(c *gin.Context) {
		userID := c.GetString("userID")
		c.JSON(http.StatusOK, gin.H{"user": userID})
	})

	// 管理员接口，只有 admin 角色才能访问
	r.GET("/admin", auth.JWTAuth("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "welcome admin"})
	})

	//r.LoadHTMLGlob("assets/html/*")
	// router.LoadHTMLGlob("templates/**/*") 多级目录
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	/*r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})*/

	r.GET("favicon.ico", func(c *gin.Context) {
		file, _ := assets.STATIC.ReadFile("static/favicon.ico")
		c.Data(
			http.StatusOK,
			"image/x-icon",
			file,
		)
	})

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
