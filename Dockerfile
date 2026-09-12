# syntax=docker/dockerfile:1
# 预编译产物镜像：二进制由外部（10.8.0.12 golang 容器）构建后放入 .build-dist/。
# 使用：先在构建机编译出 .build-dist/wb2api，再 docker compose up -d --build。
FROM alpine:3.20
RUN apk add --no-cache wget ca-certificates tzdata \
 && adduser -D -u 10001 app \
 && mkdir -p /app/auths /app/data \
 && chown -R app:app /app
USER app
WORKDIR /app
COPY .build-dist/wb2api /app/wb2api
COPY config.json /app/config.json
EXPOSE 7863
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:7863/healthz || exit 1
ENTRYPOINT ["/app/wb2api", "-config", "/app/config.json"]
