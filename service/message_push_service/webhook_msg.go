package message_push_service

import (
	"net/http"
	"net/url"
	"fmt"
	"flame_clouds/global"
	"github.com/sirupsen/logrus"
)

// WebhookMsg webhook消息推送结构体
type WebhookMsg struct {
	URL string // Webhook地址
}

// Push 实现消息推送接口
func (w WebhookMsg) Push(title string, des string) error {
	// 构建请求参数
	params := url.Values{}
	params.Add("message", title+": "+des)
	params.Add("priority", "high")
	params.Add("tags", "warning,skull")

	// 发送GET请求
	resp, err := http.Get(w.URL + "?" + params.Encode())
	if err != nil {
		logrus.Errorf("webhook推送失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("webhook推送失败, 状态码: %d", resp.StatusCode)
		return fmt.Errorf("webhook推送失败, 状态码: %d", resp.StatusCode)
	}

	logrus.Infof("webhook推送成功")
	return nil
}