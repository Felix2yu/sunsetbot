package core

import (
	"flame_clouds/config"
	"flame_clouds/config/types"
	"flame_clouds/flags"
	"flame_clouds/global"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

// ReadConfig 读取yaml配置
func ReadConfig() *config.Config {
	byteData, err := os.ReadFile(flags.Options.File)
	if err != nil {
		logrus.Fatalf("配置文件错误 %s", err)
	}
	var server config.Config
	err = yaml.Unmarshal(byteData, &server)
	if err != nil {
		logrus.Fatalf("配置文件格式错误 %s", err)
	}
	server.Monitor.Evening.EventType = types.Evening
	server.Monitor.Morning.EventType = types.Morning
	logrus.Infof("%s 配置文件加载成功", flags.Options.File)
	return &server
}

func DumpConfig() {
	byteData, err := yaml.Marshal(global.Config)
	if err != nil {
		logrus.Errorf("转换错误 %s", err)
		return
	}

	err = os.WriteFile(flags.Options.File, byteData, 0666)
	if err != nil {
		logrus.Errorf("配置文件写入失败 %s", err)
		return
	}
	logrus.Infof("配置文件写入成功")

}

