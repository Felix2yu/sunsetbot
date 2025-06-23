package cron_service

import (
	"flame_clouds/config"
	"flame_clouds/global"
	"flame_clouds/service/hsy_service"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"time"
)

var Crontab *cron.Cron

func CronService() {
	timezone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Infof("时区配置错误, %s", err)
	}

	// 看看配置文件是否正确
	if global.Config.Monitor.City == "" {
		logrus.Fatalf("未配置火烧云监控城市")
	}
	if !global.Config.Bot.Enable {
		logrus.Warnf("未配置消息通知渠道")
	}

	logrus.Infof("监控城市 %s", global.Config.Monitor.City)

	Crontab = cron.New(cron.WithSeconds(), cron.WithLocation(timezone))

	// 晚霞
	CronDate(global.Config.Monitor.Evening)

	// 朝霞
	CronDate(global.Config.Monitor.Morning)

	Crontab.Start()
}

func CronDate(e config.MonitorEvent) {

	eventString := e.EventType.String()

	if !e.Enable {
		logrus.Warnf("未配置%s火烧云监控", eventString)
		return
	}

	if e.Quality == 0 {
		logrus.Fatalf("未配置%s火烧云监控指标", eventString)
	}

	_, err := Crontab.AddFunc(e.Time, func() {
		hsy_service.GetCitySunsetData(e)
	})
	if err != nil {
		logrus.Fatalf("%s时间配置错误 %s", eventString, err)
	}

	logrus.Infof("%s监控时间 %s", eventString, e.Time)
}
