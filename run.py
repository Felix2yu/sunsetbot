import json
import requests
import schedule
import time
import datetime
import urllib.parse
import yaml
import os
import random
import re
from typing import Dict, List, Optional, Tuple
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

EVENT_MAP = {
    'TODAY_MORNING': 'rise_1',
    'TOMORROW_MORNING': 'rise_2',
    'TODAY_EVENING': 'set_1',
    'TOMORROW_EVENING': 'set_2',
}

PREDICT_MODEL_MAP = {
    'GFS': 'GFS',
    'EC': 'EC'
}

class WeatherPredictor:
    def __init__(self, config_path: str = None):
        """
        初始化天气预测器
        
        Args:
            config_path: 配置文件路径，默认为当前目录下的config.yaml
        """
        self.config = self._load_config(config_path)
        self.session = requests.Session()
        # 设置请求超时和重试策略
        self.session.timeout = (10, 30)  # 连接超时10秒，读取超时30秒
    
    def _load_config(self, config_path: str = None) -> dict:
        """加载配置文件"""
        if config_path is None:
            config_path = os.path.join(os.path.dirname(__file__), "config.yaml")
        
        try:
            with open(config_path, "r", encoding="utf-8") as f:
                return yaml.safe_load(f)
        except FileNotFoundError:
            logger.error(f"配置文件不存在: {config_path}")
            raise
        except yaml.YAMLError as e:
            logger.error(f"配置文件格式错误: {e}")
            raise
    
    def build_url(self, event: str, model: str) -> str:
        """
        构建请求URL
        
        Args:
            event: 事件类型
            model: 预测模型
            
        Returns:
            构建完成的URL
        """
        base_url = self.config["request"]["base_url"]
        params_encoded = {
            "query_id": self._generate_random_num(),
            "event": event,
            "model": model,
            "query_city": self.config["schedule"]["city"],
            "intend": "select_city",
            "event_date": "None",
            "times": "None"
        }
        
        logger.debug(f"构建参数: {params_encoded}")
        query_string = urllib.parse.urlencode(params_encoded)
        return f"{base_url}?{query_string}"
    
    @staticmethod
    def _generate_random_num() -> str:
        """生成随机数字字符串"""
        return str(random.randint(100000, 999999))
    
    def _parse_weather_data(self, content: str, target_quality: float) -> Optional[str]:
        """
        解析天气数据
        
        Args:
            content: 响应内容
            target_quality: 目标质量阈值
            
        Returns:
            解析后的天气信息字符串，如果不符合条件则返回None
        """
        try:
            json_content = json.loads(content)
            
            # 提取质量数值
            quality_str = json_content.get('tb_quality', '0.0')
            quality_match = re.search(r'\d+\.\d+', str(quality_str))
            
            if not quality_match:
                logger.warning(f"无法从质量数据中提取数值: {quality_str}")
                return None
                
            quality_num = float(quality_match.group())
            
            if quality_num < target_quality:
                logger.debug(f"质量数值 {quality_num} 低于阈值 {target_quality}")
                return None
            
            # 构建推送字符串
            push_str = f"鲜艳度：{quality_str}\n"
            push_str += f"气溶胶：{json_content.get('tb_aod', 'N/A')}\n"
            
            event_time = json_content.get('tb_event_time', '')
            if event_time:
                day_indicator = self._get_day_indicator(event_time)
                push_str += f"时间{day_indicator}：{event_time}\n"
            
            logger.debug(f"解析成功: {push_str}")
            return push_str
            
        except json.JSONDecodeError as e:
            logger.error(f"JSON解析失败: {e}, 内容: {content[:100]}...")
            return None
        except Exception as e:
            logger.error(f"解析天气数据时发生错误: {e}")
            return None
    
    def _get_day_indicator(self, event_time: str) -> str:
        """
        获取日期指示符
        
        Args:
            event_time: 事件时间字符串
            
        Returns:
            日期指示符（今天/明天）
        """
        if not event_time:
            return ''
        
        try:
            event_date = datetime.datetime.strptime(event_time[:10], '%Y-%m-%d').date()
            today = datetime.date.today()
            
            if event_date == today:
                return '(今天)'
            elif event_date == today + datetime.timedelta(days=1):
                return '(明天)'
            else:
                return f'({event_date.strftime("%m-%d")})'  # 显示具体日期
        except ValueError:
            logger.warning(f"时间格式错误: {event_time}")
            return ''
    
    def fetch_single_data(self, url: str, target_quality: float) -> Optional[str]:
        """
        获取单个URL的数据
        
        Args:
            url: 请求URL
            target_quality: 目标质量阈值
            
        Returns:
            符合条件的天气数据字符串
        """
        try:
            response = self.session.get(url, timeout=(10, 30))
            response.raise_for_status()
            content = response.text
            
            logger.info(f"请求成功: {url}")
            return self._parse_weather_data(content, target_quality)
            
        except requests.exceptions.Timeout:
            logger.error(f"请求超时: {url}")
        except requests.exceptions.RequestException as e:
            logger.error(f"请求失败: {url}, 错误: {e}")
        except Exception as e:
            logger.error(f"获取数据时发生未知错误: {e}")
        
        # 根据配置决定是否返回错误信息
        if self.config["schedule"]["push_error"]:
            return f"[失败] 请求错误: {str(e)[:100]}\n"
        return None
    
    def fetch_data(self, is_morning: bool) -> None:
        """
        获取天气数据
        
        Args:
            is_morning: 是否为早晨数据
        """
        # 确定模型配置
        section = "morning" if is_morning else "evening"
        models = self.config["schedule"][section]["model"]
        
        if not models:
            models = [PREDICT_MODEL_MAP.get("GFS")]
        
        # 构建请求URL
        urls = {}  # url -> model
        event_prefix = "MORNING" if is_morning else "EVENING"
        
        for model in models:
            # 早晨特殊处理：如果当前时间小于6点，同时获取今天和明天的朝霞
            if is_morning and datetime.datetime.now().hour < 6:
                url1 = self.build_url(EVENT_MAP[f"TODAY_{event_prefix}"], model)
                urls[url1] = model
            
            # 获取明天的朝霞/晚霞
            url_tomorrow = self.build_url(EVENT_MAP[f"TOMORROW_{event_prefix}"], model)
            urls[url_tomorrow] = model
            
            # 晚霞总是获取今天的和明天的
            if not is_morning:
                url_today = self.build_url(EVENT_MAP[f"TODAY_{event_prefix}"], model)
                urls[url_today] = model
        
        target_quality = self.config["schedule"][section]["quality"]
        event_title = "🌅 朝霞预报" if is_morning else "🌆 晚霞预报"
        
        # 并发获取数据（简化版本，可根据需要改为真正的并发）
        markdown_lines = self._build_markdown_response(urls, target_quality, event_title)
        
        if len(markdown_lines) > 2:  # 有实际数据
            push_content = "\n".join(markdown_lines)
            self.send_ntfy_notification(event_title, push_content)
        else:
            logger.info("[推送] 没有符合条件的数据")
    
    def _build_markdown_response(self, urls: Dict[str, str], 
                                target_quality: float, 
                                event_title: str) -> List[str]:
        """
        构建Markdown响应
        
        Args:
            urls: URL到模型的映射
            target_quality: 目标质量阈值
            event_title: 事件标题
            
        Returns:
            Markdown行列表
        """
        markdown_lines = []
        city = self.config["schedule"]["city"]
        
        markdown_lines.extend([
            f"**城市：** {city}",
            f"**鲜艳度阈值：** {target_quality}",
            ""
        ])
        
        # 并行处理（这里简化为顺序处理，可根据需要改为并发）
        for url, model in urls.items():
            result = self.fetch_single_data(url, target_quality)
            if result:
                markdown_lines.append(f"**模型：{model}**")
                for line in result.strip().split('\n'):
                    if line.strip():  # 跳过空行
                        markdown_lines.append(f"- {line}")
                markdown_lines.append("")  # 分隔
        
        return markdown_lines
    
    def send_ntfy_notification(self, title: str, content: str) -> None:
        """
        发送ntfy通知
        
        Args:
            title: 通知标题
            content: 通知内容
        """
        push_enable = self.config["push"]["enable"]
        if not push_enable:
            logger.info("[推送已关闭]")
            return
        
        ntfy_server = self.config["push"].get("ntfy_server", "https://ntfy.sh").rstrip('/')
        ntfy_topic = self.config["push"].get("ntfy_topic")
        
        if not ntfy_topic:
            logger.error("[推送失败] 配置中未设置 ntfy_topic")
            return
        
        url = f"{ntfy_server}/{ntfy_topic}"
        
        headers = {"Markdown": "yes"}
        token = self.config["push"].get("ntfy_token")
        if token:
            headers["Authorization"] = f"Bearer {token}"
        
        message = f"# {title}\n\n{content}"
        
        try:
            response = self.session.post(url, data=message.encode('utf-8'), headers=headers)
            response.raise_for_status()
            logger.info(f"[推送成功] ntfy 通知已发送到 {url}")
        except Exception as e:
            logger.error(f"[推送失败] {e}")


def main():
    """主函数"""
    logger.info(f"[启动] {datetime.datetime.now()}")
    
    try:
        predictor = WeatherPredictor()
    except Exception as e:
        logger.error(f"初始化失败: {e}")
        return
    
    # 任务调度配置
    config = predictor.config
    morning_task_enable = config["schedule"]["morning"]["enable"]
    evening_task_enable = config["schedule"]["evening"]["enable"]
    push_enable = config["push"]["enable"]
    
    # 配置早晨任务
    if morning_task_enable:
        morning_times = config["schedule"]["morning"]["time"]
        if morning_times:
            for run_time in morning_times:
                run_time_str = str(run_time).strip()
                logger.info(f"[启动] 朝霞任务将每天 {run_time_str} 执行")
                schedule.every().day.at(run_time_str).do(predictor.fetch_data, True)
    
    # 配置晚上任务
    if evening_task_enable:
        evening_times = config["schedule"]["evening"]["time"]
        if evening_times:
            for run_time in evening_times:
                run_time_str = str(run_time).strip()
                logger.info(f"[启动] 晚霞任务将每天 {run_time_str} 执行")
                schedule.every().day.at(run_time_str).do(predictor.fetch_data, False)
    
    logger.info(
        f"[启动] 朝霞任务：{morning_task_enable} "
        f"晚霞任务：{evening_task_enable} "
        f"微信通知: {push_enable} "
        f"推送异常：{config['schedule']['push_error']}"
    )
    
    # 启动时发送测试消息
    if config["schedule"]["send_test_on_start"]:
        predictor.send_ntfy_notification("服务启动测试", "服务已启动，这是一条测试消息")
    
    # 运行调度器
    while True:
        schedule.run_pending()
        time.sleep(1)


if __name__ == "__main__":
    main()
