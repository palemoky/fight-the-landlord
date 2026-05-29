# 定义构建参数，设置一个默认值（以防本地直接 docker build 时没有传参）
ARG GO_VERSION=1.26

# 构建阶段（dev 变体带 shell/编译器/git，用于 CI 构建）
FROM dhi.io/golang:${GO_VERSION}-dev AS builder

# dev 变体默认非 root，构建阶段切到 root 以便写入缓存目录
USER root

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建服务端二进制（CGO 关闭，产出静态二进制）
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /server ./cmd/server

# 运行阶段：DHI static 为 distroless 风格，自带 ca-certificates，默认非 root
# static 无 latest 标签，需固定到带日期的 tag（catalog 中查最新）
FROM dhi.io/static:20260413-alpine3.23

WORKDIR /app

# 时区设为 UTC（Go 内建，无需 zoneinfo / tzdata）
ENV TZ=UTC

# 从构建阶段复制二进制文件
COPY --from=builder /server /app/server

# 复制配置文件
COPY config.yaml /app/config.yaml

# 暴露端口
EXPOSE 1780

# 注意：static 运行镜像无 shell / wget，原 HEALTHCHECK 已移除。
# 若需容器级健康检查，请在编排层（compose/k8s）配置，或给二进制增加 -health 自检子命令。

# 运行服务
CMD ["/app/server"]
