# 使用官方 Python 基础镜像
FROM python:3-slim

# 将构建环境下的文件OR目录, 复制到镜像中的/code目录下, 
# ADD . /code

# 设置工作目录
WORKDIR /app

# 复制依赖文件并安装
COPY requirements.txt .
RUN pip install -r requirements.txt
 
# 复制项目代码
COPY config.yaml /app/
COPY run.py /app/

# RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
# RUN echo 'Asia/Shanghai' > /etc/timezone

ENV TIME_ZONE=Asia/Shanghai 
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone

# 启动程序
CMD ["python", "run.py"]
