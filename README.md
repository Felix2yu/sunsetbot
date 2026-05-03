## 这是一个python写的朝霞晚霞预警脚本程序
用户可根据config.yaml文件配置每天查询火烧云的时间和预期质量，最后通过ntfy推送信息

由sunsetbot.top提供接口

## 消息推送
使用ntfy推送信息

官方ntfy地址：https://ntfy.sh/

也可自建部署本地服务。

## 配置
```yaml
request:
  base_url: "https://sunsetbot.top/"  #不用动

# 信息推送
push:
  enable: true
  ntfy_server: "https://ntfy.sh"      # 可自定义自建服务器地址
  ntfy_topic: "Weather"               # 替换为实际主题

schedule:
  city: "江苏省-苏州"
  send_test_on_start: false           # 是否在启动的时候推送测试通知
  push_error: false                   # 请求错误是否推送
  # 朝霞
  morning:  
    enable: true
    quality: 0.3                      # 质量
    time: ["18:00","23:00"]           # 多个时间用英文逗号隔开
    model: ["GFS","EC"]               #"GFS","EC"  多个模式用英文逗号隔开
  
  # 晚霞
  evening: 
    enable: true
    quality: 0.2
    time: ["08:00", "11:00", "16:00"]
    model: ["GFS","EC"] 
```

如果使用docker，请映射config.yaml ```docker-compose.yaml```仅供参考

### Docker命令
打包镜像：
docker build . -t sunsetbot

查看镜像：
docker image ls

导出镜像：（不要通过镜像id导出，不然导入后看不到导入的镜像）
docker save -o sunsetbot.tar sunsetbot
