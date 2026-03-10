package logger

/*import "go.uber.org/zap"

var Log *zap.Logger

func Init() {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	//defer Log.Sync()
}*/

import (
	"os"

	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func Init() {

	cfg := config.Config

	logFile := cfg.GetString("log.file")

	hook := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    cfg.GetInt("log.maxSize"),    //文件大小(MB)
		MaxBackups: cfg.GetInt("log.maxBackups"), //保留数量
		MaxAge:     cfg.GetInt("log.maxAge"),     //保留天数
		Compress:   cfg.GetBool("log.compress"),  //gzip压缩
	}

	writeSyncer := zapcore.AddSync(hook)

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// 设置系统的日志级别
	level := parseLevel(cfg.GetString("log.level"))

	// 3. 控制台输出
	consoleWriter := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(
		encoder,
		//writeSyncer,
		zapcore.NewMultiWriteSyncer(consoleWriter, writeSyncer), // 多日志
		level,
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {

		}
	}(Log)
}

func parseLevel(level string) zapcore.Level {

	switch level {

	case "debug":
		return zap.DebugLevel

	case "warn":
		return zap.WarnLevel

	case "error":
		return zap.ErrorLevel

	default:
		return zap.InfoLevel
	}
}
