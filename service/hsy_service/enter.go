package hsy_service

import (
	"encoding/json"
	"flame_clouds/config"
	"flame_clouds/global"
	"flame_clouds/service/message_push_service"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type SunsetBotReq struct {
	City  string  `json:"city"`
	Aod   float64 `json:"aod"`
	Event string  `json:"event"` // set_1:今天日落, rise_2:明天日出
}

type SunsetBotResponse struct {
	ImgHref     string `json:"img_href"`
	ImgSummary  string `json:"img_summary"`
	PlaceHolder string `json:"place_holder"`
	QueryId     string `json:"query_id"`
	Status      string `json:"status"`
	TbAod       string `json:"tb_aod"`
	TbEventTime string `json:"tb_event_time"` // 事件时间
	TbQuality   string `json:"tb_quality"`    // 火烧云指标
}

// 生成随机查询ID
func generateQueryID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(r.Intn(10000000) + 1)
}

// GetSunsetData 获取日落/日出数据
func GetSunsetData(req SunsetBotReq) (*SunsetBotResponse, error) {
	queryID := generateQueryID()
	baseURL := "https://sunsetbot.top/"

	params := url.Values{}
	params.Add("query_id", queryID)
	params.Add("intend", "select_city")
	params.Add("query_city", req.City)
	params.Add("event_date", "None")
	params.Add("event", req.Event)
	params.Add("times", "None")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var data SunsetBotResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}
	return &data, nil
}

// GetCitySunsetData 获取指定城市的天气数据
func GetCitySunsetData(e config.MonitorEvent) {
	t, err := GetSunsetData(SunsetBotReq{City: global.Config.Monitor.City, Event: e.EventType.Params(), Aod: e.Quality})
	if err != nil {
		logrus.Errorf("请求错误 %s", err)
		return
	}
	checkAndNotify(t, e)
}

// 解析火烧云指标
func parseQuality(qualityStr string) (float64, error) {
	// 去除HTML标签和额外内容
	cleanStr := strings.Split(qualityStr, "<")[0]
	cleanStr = strings.TrimSpace(cleanStr)
	return strconv.ParseFloat(cleanStr, 64)
}

// 检查并处理火烧云指标
func checkAndNotify(data *SunsetBotResponse, e config.MonitorEvent) {
	quality, err := parseQuality(data.TbQuality)
	if err != nil {
		logrus.Errorf("解析火烧云质量失败: %s", err)
		return
	}

	logrus.Infof("城市: %s, 事件: %s, 质量: %.2f", global.Config.Monitor.City, e.EventType.String(), quality)

	if quality < e.Quality {
		logrus.Warnf("火烧云指标未达到阈值")
		return
	}

	// 构建消息内容
	message := fmt.Sprintf(
		"【火烧云预警】城市: %s\n事件: %s\n时间: %s\n火烧云质量: %.2f\n气溶胶指标: %s\n满足拍摄条件!",
		global.Config.Monitor.City,
		e.EventType.String(),
		data.TbEventTime,
		quality,
		data.TbAod,
	)
	message = strings.ReplaceAll(message, "<br>", "")
	logrus.Infof(message)

	if !global.Config.Bot.Enable {
		logrus.Infof("未配置消息推送渠道")
		return
	}

	// 去请求图片数据
	if global.Config.Monitor.Map.Enable {
		response, err1 := GetSunsetMapData(MapReq{
			Region: global.Config.Monitor.Map.Region,
			Event:  e.EventType.Params(),
		})
		if err1 == nil {
			message += fmt.Sprintf("\n![](%s)", "https://sunsetbot.top"+response.MapImgSrc)
		} else {
			logrus.Errorf("请求火烧云地图数据失败 %s", err1)
		}
	}

	title := fmt.Sprintf("[%s] %s预警 质量:%.2f", global.Config.Monitor.City, e.EventType.String(), quality)

	// 消息推送
	bot := message_push_service.NewMessage(global.Config.Bot.Target)
	if bot == nil {
		return
	}
	err = bot.Push(title, message)
	if err != nil {
		logrus.Errorf("消息推送失败 %s", err)
	}

}
