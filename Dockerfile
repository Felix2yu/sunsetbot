# 多阶段构建：第一阶段 - 构建阶段
FROM python:3.14-alpine AS builder

WORKDIR /build

# 复制依赖文件
COPY requirements.txt .

# 安装依赖到本地目录
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# 第二阶段 - 运行阶段
FROM python:3.14-alpine

WORKDIR /app

# 复制依赖
COPY --from=builder /install /usr/local

# 复制项目代码
COPY config.yaml .
COPY run.py .

# 设置时区
ENV TIME_ZONE=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime \
    && echo $TIME_ZONE > /etc/timezone

# 启动程序
CMD ["python", "run.py"]
