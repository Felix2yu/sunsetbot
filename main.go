package main

import (
	"flame_clouds/core"
	"flame_clouds/flags"
	"flame_clouds/global"
	"flame_clouds/service/cron_service"
)

func main() {
	// 读取配置文件
	global.Config = core.ReadConfig()
	// 日志
	core.InitLogger()

	// 命令行参数绑定
	flags.Run()

	// 定时任务
	cron_service.CronService()

	select {}
}

