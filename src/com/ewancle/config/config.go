package config

/*import "github.com/spf13/viper"

var Config *viper.Viper

func Init() {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}

	Config = v
}*/

import (
	"bytes"
	"os"
	"regexp"

	configs_embed "github.com/mpv945/openwrt-custom-tools-gin/configs"
	"github.com/spf13/viper"
)

var Config *viper.Viper

func Init() {

	/*v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}*/

	v := viper.New()

	v.SetConfigType("yaml")

	// windows 打包
	//      $env:CGO_ENABLED="0"; $env:GOOS="windows"; $env:GOARCH="amd64"; go build -trimpath -ldflags "-s -w" -o build/app.exe ./src/com/ewancle
	// 如果打包成可执行文件，可以优先在可执行文件同层级目录放置 configs/config.yaml
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	err := v.ReadInConfig()

	if err != nil {
		// 如果外部配置不存在，使用 embed.FS
		data, readErr := configs_embed.CFS.ReadFile("config.yaml")
		if readErr != nil {
			panic(readErr)
		}

		// viper 从 bytes.Buffer 读取
		err = v.ReadConfig(bytes.NewBuffer(data))
		if err != nil {
			panic(err)
		}
	}

	// 解析 ${ENV:default}
	resolveEnv(v)

	Config = v
}

// 这样就支持：${LOG_LEVEL:info}
func resolveEnv(v *viper.Viper) {

	re := regexp.MustCompile(`\$\{(\w+):?(.*?)\}`)

	for _, key := range v.AllKeys() {

		val := v.GetString(key)

		matches := re.FindStringSubmatch(val)

		if len(matches) == 3 {

			env := os.Getenv(matches[1])

			if env != "" {
				v.Set(key, env)
			} else {
				v.Set(key, matches[2])
			}

		}
	}
}
