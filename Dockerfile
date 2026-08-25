# 运行时镜像：二进制由 CI 矩阵预编译并下载到 bin/ 后拼装
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates gosu \
    && addgroup -g 1000 appuser \
    && adduser -D -u 1000 -G appuser appuser

COPY --chmod=755 bin/liuxia /app/liuxia
COPY templates /app/templates
COPY static /app/static
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENV TZ=Asia/Shanghai \
    PUID=1000 \
    PGID=1000

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
