## 朝霞晚霞预警脚本程序

用户可根据`config.yaml`文件配置每天查询火烧云的时间和预期质量，最后通过ntfy推送信息

由 sunsetbot.top 提供接口

## 配置

```YAML
request:
  base_url: "https://sunsetbot.top/"      # 不用动

# 信息推送
push:
  enable: true
  ntfy_server: "https://ntfy.sh"          # 可自定义自建服务器地址
  ntfy_topic: "Weather"                   # 替换为实际主题

schedule:
  city: "江苏省-苏州"
  send_test_on_start: false               # 是否在启动的时候推送测试通知
  push_error: true                        # 请求错误是否推送
  # 朝霞
  morning:  
    enable: true
    time: ["18:00","00:00"]               # 多个时间用英文逗号隔开
    model: ["GFS","EC"]                   # 当前支持GFS、EC，多个模式之间用英文逗号隔开
  
  # 晚霞
  evening: 
    enable: true
    time: ["08:00", "11:30", "16:00"]
    model: ["GFS","EC"]
```

## 消息推送

使用ntfy推送信息，也可自建部署本地服务。

官方ntfy地址：<https://ntfy.sh/>

页面上新建Topic后填入配置文件中，暂不支持需要验证身份的Topic。

### 通知等级

Ntfy 通知等级对应关系：

- 过滤阈值： < 0.2 的数据会被过滤掉，不通知
- 0.2 - 0.4 → 等级 1
- 0.4 - 0.6 → 等级 2
- 0.6 - 0.8 → 等级 3
- 0.8 - 1.0 → 等级 4
- 1.0 及以上 → 等级 5

ntfy消息中质量、气溶胶数值较优秀时会加粗标记。

![](.img/snapshot.jpg)
