# ============ 阶段1: 构建前端 ============
FROM node:18-alpine AS frontend

WORKDIR /web

# 使用国内镜像源加速（可根据需要修改）
COPY web/package*.json ./
RUN npm install --registry=https://registry.npmmirror.com

# 复制前端代码并构建
COPY web/ .
RUN npm run build

# ============ 阶段2: 构建后端 ============
FROM golang:1.21-alpine AS backend

WORKDIR /app

# 设置 Go 代理并下载依赖
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

# 复制后端代码
COPY . .

# 复制前端构建产物
COPY --from=frontend /web/dist ./web/dist

# 编译（静态链接，可在任何 Linux 运行）
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=$(cat VERSION 2>/dev/null || echo dev) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/server ./cmd/server

# ============ 阶段3: 运行时 ============
FROM alpine:3.19

# 安装必要的运行时依赖
# - ca-certificates: HTTPS 请求
# - tzdata: 时区支持
# - wget: 健康检查
RUN apk add --no-cache ca-certificates tzdata wget

# 创建非 root 用户运行
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 复制二进制文件和前端资源
COPY --from=backend --chown=appuser:appgroup /app/server .
COPY --from=backend --chown=appuser:appgroup /app/web ./web
COPY --from=backend --chown=appuser:appgroup /app/config.yaml .

# 创建数据目录
RUN mkdir -p /data/cache && \
    chown -R appuser:appgroup /data

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 18900

# 健康检查（使用不需要认证的接口）
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:18900/api/health || exit 1

# 环境变量默认值
ENV CONFIG_PATH=/app/config.yaml
ENV GIN_MODE=release
ENV TZ=Asia/Shanghai

# 入口
CMD ["./server"]
