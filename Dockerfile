# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建服务端二进制
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /server ./cmd/server

# 运行阶段
FROM alpine:3

WORKDIR /app

# 安装 ca-certificates（如果需要 HTTPS）
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /server /app/server

# 复制配置文件
COPY config.yaml /app/config.yaml

# 暴露端口
EXPOSE 1780

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:1780/health || exit 1

# 运行服务
CMD ["/app/server"]
