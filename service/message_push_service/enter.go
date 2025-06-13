package message_push_service

import (
	"flame_clouds/config/types"
	"flame_clouds/global"
	"github.com/sirupsen/logrus"
)

// MessagePushInterface 拿到数据之后，根据配置文件中填写的目标，进行推送
type MessagePushInterface interface {
	Push(title string, des string) error
}

func NewMessage(t types.BotTargetType) MessagePushInterface {
	switch t {
	case types.FtBot: // 方糖
		return FtMsg{
			Key: global.Config.Bot.SendKey,
		}
	case types.WebhookBot: // Webhook
		return WebhookMsg{
			URL: global.Config.Bot.WebhookURL,
		}
	default:
		logrus.Errorf("消息推送配置错误")
		return nil
	}
}
