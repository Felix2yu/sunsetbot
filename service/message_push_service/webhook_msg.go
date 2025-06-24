package message_push_service

import (
	"net/http"
	"fmt"
	"io"
	"strings"
	"github.com/sirupsen/logrus"
)

// WebhookMsg webhook消息推送结构体
type WebhookMsg struct {
	URL string // Webhook地址
}

// Push 实现消息推送接口
func (w WebhookMsg) Push(title string, des string) error {
	// 构建请求体
	bodyContent := fmt.Sprintf("%s", des)
	logrus.Infof("webhook推送请求体: %s", bodyContent)

	// 创建POST请求
	req, err := http.NewRequest("POST", w.URL, strings.NewReader(bodyContent))
	req.Header.Set("Title", title)
	if err != nil {
		logrus.Errorf("创建webhook请求失败: %v", err)
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Markdown", "yes")
	req.Header.Set("Priority", "high")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("webhook推送请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("webhook响应读取失败: %v", err)
		return err
	}
	responseBody := string(body)

	logrus.Infof("webhook推送响应: 状态码=%d, 内容=%s", resp.StatusCode, responseBody)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook推送失败, 状态码: %d, 响应内容: %s", resp.StatusCode, responseBody)
	}

	logrus.Infof("webhook推送成功")
	return nil
}
