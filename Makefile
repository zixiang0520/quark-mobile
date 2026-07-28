.PHONY: dev dev-web dev-api build build-web build-api \
        docker docker-build docker-build-multi docker-run docker-stop \
        docker-push docker-release \
        clean help

# ============ 变量 ============
APP_NAME := quark-mobile
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")
REGISTRY ?= docker.io
IMAGE_NAME := $(REGISTRY)/$(APP_NAME)
IMAGE_TAG := $(VERSION)

# ============ 开发模式 ============
dev: dev-api
	@echo "使用 'make dev-web' 启动前端, 'make dev-api' 启动后端"

dev-api:
	go run cmd/server/main.go

dev-web:
	cd web && npm run dev

# ============ 本地构建 ============
build: build-web build-api

build-web:
	cd web && npm run build

build-api:
	CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ============ Docker 构建 ============
# 单架构构建（默认当前架构）
docker: docker-build

docker-build:
	@echo "📦 构建 Docker 镜像: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker build \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest \
		.
	@echo "✅ 构建完成: $(IMAGE_NAME):$(IMAGE_TAG)"

# 多架构构建 (amd64 + arm64)
docker-build-multi:
	@echo "📦 构建多架构 Docker 镜像: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest \
		--load \
		.
	@echo "✅ 多架构构建完成"

# ============ Docker 运行 ============
docker-run:
	@echo "🚀 启动容器..."
	docker run -d \
		-p 18900:18900 \
		-v $(shell pwd)/config.yaml:/app/config.yaml:ro \
		-v quark-mobile-data:/data \
		--name $(APP_NAME) \
		--restart unless-stopped \
		$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "✅ 容器已启动: http://localhost:18900"

docker-run-env:
	@echo "🚀 启动容器（使用环境变量配置）..."
	docker run -d \
		-p 18900:18900 \
		-e OL_BASE_URL=$(OL_BASE_URL) \
		-e OL_USERNAME=$(OL_USERNAME) \
		-e OL_PASSWORD=$(OL_PASSWORD) \
		-e OL_MOUNT_QUARK=$(OL_MOUNT_QUARK) \
		-e OL_MOUNT_MOBILE=$(OL_MOUNT_MOBILE) \
		-v quark-mobile-data:/data \
		--name $(APP_NAME) \
		--restart unless-stopped \
		$(IMAGE_NAME):$(IMAGE_TAG)

docker-stop:
	docker stop $(APP_NAME) 2>/dev/null || true
	docker rm $(APP_NAME) 2>/dev/null || true
	@echo "✅ 容器已停止"

# ============ Docker Compose ============
compose-up:
	docker-compose up -d

compose-down:
	docker-compose down

compose-logs:
	docker-compose logs -f

# ============ 镜像发布 ============
# 发布到远程仓库
docker-push:
	@echo "📤 推送镜像到 $(IMAGE_NAME):$(IMAGE_TAG)"
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_NAME):latest
	@echo "✅ 推送完成"

# 发布多架构镜像
docker-push-multi:
	@echo "📤 推送多架构镜像到 $(IMAGE_NAME):$(IMAGE_TAG)"
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest \
		--push \
		.
	@echo "✅ 多架构推送完成"

# 一键发布（构建+推送）
docker-release: docker-build docker-push

# ============ 镜像导出 ============
docker-export:
	@echo "💾 导出镜像到 $(APP_NAME)-$(IMAGE_TAG).tar"
	docker save $(IMAGE_NAME):$(IMAGE_TAG) -o $(APP_NAME)-$(IMAGE_TAG).tar
	@echo "✅ 导出完成, 可用 'docker load -i $(APP_NAME)-$(IMAGE_TAG).tar' 导入"

# ============ 清理 ============
clean:
	rm -rf web/dist
	rm -f server
	rm -rf data/cache/*
	rm -f *.tar

# ============ 辅助 ============
docker-image-clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
	docker rmi $(IMAGE_NAME):latest 2>/dev/null || true
	docker image prune -f

help:
	@echo ""
	@echo "╔══════════════════════════════════════════════════╗"
	@echo "║     quark-mobile 网盘互传工具 - Makefile 帮助    ║"
	@echo "╚══════════════════════════════════════════════════╝"
	@echo ""
	@echo "📦 构建相关:"
	@echo "  make dev-api              启动后端开发服务器"
	@echo "  make dev-web              启动前端开发服务器"
	@echo "  make build                构建前端和后端"
	@echo ""
	@echo "🐳 Docker 相关:"
	@echo "  make docker               构建 Docker 镜像"
	@echo "  make docker-build-multi   构建多架构镜像(amd64+arm64)"
	@echo "  make docker-run           运行容器(使用 config.yaml)"
	@echo "  make docker-run-env       运行容器(使用环境变量)"
	@echo "  make docker-stop          停止容器"
	@echo "  make compose-up           使用 docker-compose 启动"
	@echo ""
	@echo "📤 发布相关:"
	@echo "  make docker-push          推送镜像到远程仓库"
	@echo "  make docker-push-multi    推送多架构镜像"
	@echo "  make docker-release       一键构建+推送"
	@echo "  make docker-export        导出镜像为 tar 文件"
	@echo ""
	@echo "🔧 变量:"
	@echo "  VERSION       版本号 (默认: 0.1.0)"
	@echo "  REGISTRY      镜像仓库 (默认: docker.io)"
	@echo ""
	@echo "示例:"
	@echo "  make docker REGISTRY=ghcr.io VERSION=1.0.0"
	@echo "  make docker-release REGISTRY=ghcr.io VERSION=1.0.0"
	@echo ""
