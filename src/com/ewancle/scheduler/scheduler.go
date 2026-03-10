package scheduler

import (
	"fmt"
	"time"

	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func Start() {

	//c := cron.New(cron.WithSeconds())
	c := cron.New(
		cron.WithSeconds(),
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger),
			cron.Recover(cron.DefaultLogger),
		),
	)

	_, err0 := c.AddFunc("@every 10s", func() {
		fmt.Println("定时任务执行:", time.Now())
	})
	if err0 != nil {
		return
	}

	// 每分钟执行
	_, err := c.AddFunc("0 */1 * * * *", func() {
		logger.Log.Info("执行清理任务")
	})
	if err != nil {
		logger.Log.Error(err.Error())
	}

	// 每10秒执行
	_, err1 := c.AddFunc("*/10 * * * * *", func() {
		// 延迟10秒执行
		time.Sleep(10 * time.Second)
		fmt.Println("任务执行:", time.Now())
		logger.Log.Info("任务执行:", zap.Time("seconds", time.Now()))
	})
	if err1 != nil {
		logger.Log.Error(err1.Error())
	}

	// 延迟任务
	/*delay := 30 * time.Second
	runTime := time.Now().Add(delay)

	spec := fmt.Sprintf(
		"%d %d %d %d %d *",
		runTime.Second(),
		runTime.Minute(),
		runTime.Hour(),
		runTime.Day(),
		int(runTime.Month()),
	)

	c.AddFunc(spec, func() {
		fmt.Println("延迟任务执行:", time.Now())
	})*/
	c.Start()
}
