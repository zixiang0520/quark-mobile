# ☁️ 夸克 ↔ 移动云盘传输工具

基于 OpenList 驱动实现夸克网盘和移动云盘文件互传的工具，支持 SHA256 秒传，通过 Docker 一键部署。

## ✨ 功能特性

- **🚀 极速传输**：通过 SHA256 哈希校验，支持秒传（文件已存在于目标网盘时）
- **🔄 双向传输**：支持夸克网盘 ↔ 移动云盘 之间的文件互传
- **🔐 安全加密**：密码使用 AES-GCM 加密存储，会话管理防止未授权访问
- **🌐 Web 界面**：美观的网页操作界面，随时随地管理传输任务
- **🐳 Docker 部署**：容器化设计，一键部署，支持 x86/ARM 架构
- **🔧 OpenList 驱动**：基于 OpenList REST API，统一管理多网盘

## 🏗️ 架构设计

```
┌─────────────┐     ┌─────────────┐
│  夸克网盘    │     │  移动云盘    │
└──────┬──────┘     └──────┬──────┘
       │                    │
       ▼                    ▼
┌─────────────────────────────────┐
│          OpenList 服务           │
│      (统一挂载点管理)             │
└─────────────────┬───────────────┘
                  │
                  ▼
┌─────────────────────────────────┐
│     Quark-Mobile 传输服务        │
│  ┌─────────────────────────┐   │
│  │  SHA256 校验 / 秒传逻辑  │   │
│  └─────────────────────────┘   │
│  ┌─────────────────────────┐   │
│  │  任务队列 / 进度追踪     │   │
│  └─────────────────────────┘   │
└─────────────────┬───────────────┘
                  │
                  ▼
         Web 管理界面
```

## 🚀 快速开始

### 方式一：Docker Compose（推荐）

1. 创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  quark-mobile:
    image: ghcr.io/zixiang0520/quark-mobile:latest
    container_name: quark-mobile
    
    # 端口映射
    ports:
      - "18900:18900"
    
    # 数据持久化
    volumes:
      - quark-mobile-data:/data
    
    # 环境变量配置
    environment:
      - CONFIG_PATH=/app/config.yaml
      - GIN_MODE=release
      - TZ=Asia/Shanghai
      # OpenList 连接配置（修改为你的实际配置）
      - OL_BASE_URL=http://your-openlist:5244
      - OL_USERNAME=admin
      - OL_PASSWORD=your_openlist_password
      # 挂载路径
      - OL_MOUNT_QUARK=/quark
      - OL_MOUNT_MOBILE=/mobile
    
    restart: unless-stopped
    
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:18900/api/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    
    networks:
      - quark-mobile-net

networks:
  quark-mobile-net:
    driver: bridge

volumes:
  quark-mobile-data:
    driver: local
```

2. 启动服务：

```bash
docker compose up -d
```

3. 查看日志：

```bash
docker compose logs -f
```

4. 停止服务：

```bash
docker compose down
```

### 方式二：Docker Run

```bash
# 拉取镜像
docker pull ghcr.io/zixiang0520/quark-mobile:latest

# 运行容器
docker run -d \
  --name quark-mobile \
  -p 18900:18900 \
  -v quark-mobile-data:/data \
  -e OL_BASE_URL=http://your-openlist:5244 \
  -e OL_USERNAME=admin \
  -e OL_PASSWORD=your_password \
  --restart unless-stopped \
  ghcr.io/zixiang0520/quark-mobile:latest
```

### 方式三：本地构建

```bash
# 克隆仓库
git clone https://github.com/zixiang0520/quark-mobile.git
cd quark-mobile

# 构建并启动
docker compose up -d --build
```

## ⚙️ 配置说明

### 初始管理员密码

首次启动后，使用默认密码登录：
- **用户名**：无（只需输入密码）
- **默认密码**：`admin123`

登录后请立即在「⚙️ 配置」页面修改密码。

### OpenList 绑定

1. 登录后点击右上角「⚙️ 配置」按钮
2. 在「OpenList 连接设置」中填写：
   - **OpenList 地址**：例如 `http://your-openlist:5244`
   - **登录用户名**：OpenList 的管理员用户名
   - **登录密码**：OpenList 的管理员密码
   - **夸克网盘挂载路径**：例如 `/quark`（OpenList 中夸克网盘的挂载点）
   - **移动云盘挂载路径**：例如 `/mobile`（OpenList 中移动云盘的挂载点）
3. 点击「🔍 测试连接」验证配置
4. 点击「💾 保存配置」保存（密码将使用 AES-GCM 加密存储）

### 环境变量配置

可通过环境变量预设 OpenList 连接信息：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `OL_BASE_URL` | OpenList 服务地址 | `http://localhost:5244` |
| `OL_USERNAME` | OpenList 登录用户名 | `admin` |
| `OL_PASSWORD` | OpenList 登录密码 | 空 |
| `OL_MOUNT_QUARK` | 夸克网盘挂载路径 | `/quark` |
| `OL_MOUNT_MOBILE` | 移动云盘挂载路径 | `/mobile` |
| `CONFIG_PATH` | 配置文件路径 | `/app/config.yaml` |
| `GIN_MODE` | 运行模式（debug/release） | `release` |
| `TZ` | 时区设置 | `Asia/Shanghai` |

### 配置文件说明

配置文件 `config.yaml` 示例：

```yaml
server:
  port: 18900
  mode: release

openlist:
  base_url: http://your-openlist:5244
  username: admin
  password: ""  # 加密存储，留空则不修改
  mounts:
    quark: /quark
    mobile: /mobile

transfer:
  max_concurrent: 4
  cache_dir: /data/cache
  timeout: 60
```

## 📖 使用指南

### 基本传输流程

1. **登录系统**：访问 `http://your-server:18900`，输入管理员密码
2. **配置 OpenList**：在「⚙️ 配置」页面绑定 OpenList 服务
3. **选择源文件**：在左侧「源网盘」选择要传输的文件
4. **设置目标路径**：在右侧「目标网盘」选择保存位置
5. **开始传输**：点击「📤 创建传输任务」区域的「开始传输」按钮
6. **查看进度**：在「📋 传输任务」区域查看实时进度

### 传输策略

1. **秒传检查**：系统会计算源文件的 SHA256 哈希值
2. **存在性检测**：查询目标网盘是否已存在相同 SHA256 的文件
3. **秒传执行**：如果目标文件存在，直接通过 OpenList Copy 接口秒传
4. **下载上传**：如果目标文件不存在，先下载再上传

## 🔌 API 接口

### 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| POST | `/api/login` | 管理员登录 |
| POST | `/api/logout` | 管理员登出 |
| POST | `/api/settings/test` | 测试 OpenList 连接 |

### 认证接口（需要 Session）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/settings` | 获取当前配置 |
| POST | `/api/settings` | 保存配置 |
| POST | `/api/settings/password` | 修改管理员密码 |
| GET | `/api/drivers` | 获取已注册的驱动列表 |
| GET | `/api/files/:driver` | 浏览网盘文件 |
| POST | `/api/transfer` | 创建传输任务 |
| GET | `/api/tasks` | 获取任务列表 |
| GET | `/api/tasks/:id` | 获取单个任务详情 |
| DELETE | `/api/tasks/:id` | 取消传输任务 |

### 认证方式

登录后，服务器会返回 `session_id`，后续请求需要：
- 通过 Cookie 传递：`session_id`
- 或通过 Header 传递：`X-Session-ID: <session_id>`

## 🔒 安全说明

### 加密机制

- **管理员密码**：使用 SHA-256 哈希存储
- **OpenList 密码**：使用 AES-256-GCM 加密存储
- **Session 管理**：24 小时有效期的随机 Token

### 权限控制

- 配置页面所有 API 都需要登录认证
- 登录接口有速率限制（防止暴力破解）
- 非 root 用户运行（容器内）

### 生产环境建议

1. 修改默认管理员密码
2. 使用 HTTPS 反向代理（如 Nginx、Caddy）
3. 限制访问 IP 白名单
4. 定期备份 `/data` 目录

## 📁 目录结构

```
quark-mobile/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── api/                     # API 路由和处理器
│   ├── config/                  # 配置管理、加密、Session
│   ├── driver/                  # 网盘驱动（OpenList）
│   ├── port/                    # 驱动接口定义
│   ├── model/                   # 数据模型
│   ├── service/                 # 业务逻辑（传输服务）
│   └── task/                    # 任务管理
├── web/                         # 前端代码
│   └── src/
│       ├── components/
│       │   ├── LoginPage.jsx    # 登录页面
│       │   └── BindingPage.jsx  # 绑定配置页面
│       └── App.jsx
├── Dockerfile                   # Docker 构建文件
├── docker-compose.yml           # Docker Compose 配置
├── config.yaml                  # 默认配置文件
└── README.md
```

## 🛠️ 技术栈

### 后端
- **语言**：Go 1.21
- **框架**：Gin（HTTP 框架）
- **加密**：AES-256-GCM、SHA-256

### 前端
- **框架**：React 18
- **构建工具**：Vite 5
- **样式**：原生 CSS

### 部署
- **容器化**：Docker（多阶段构建）
- **编排**：Docker Compose
- **镜像仓库**：GitHub Container Registry

## ❓ 常见问题

### Q: 如何重置管理员密码？
A: 删除 `/data/.admin_pass` 文件，重启容器后将恢复默认密码 `admin123`。

### Q: OpenList 连接测试失败？
A: 请检查：
1. OpenList 服务是否正常运行
2. 地址是否正确（包含端口号）
3. 用户名密码是否正确
4. 网络是否可达（容器内可使用 `docker exec -it quark-mobile ping your-openlist`）

### Q: 秒传功能如何工作？
A: 系统会计算文件的 SHA256 哈希值，如果目标网盘已存在相同哈希的文件，OpenList 会直接在服务端复制，无需下载上传。

### Q: 支持哪些网盘？
A: 本工具基于 OpenList，支持 OpenList 能挂载的所有网盘。当前已测试：
- 夸克网盘
- 移动云盘
- 阿里云盘
- 百度网盘
- 等等...

### Q: 如何自定义挂载路径？
A: 在 OpenList 中挂载网盘后，记下挂载路径（如 `/quark`、`/mobile`），在「⚙️ 配置」页面填入即可。

### Q: 数据会存在哪里？
A: 
- 配置信息存储在 `/data/config.yaml`
- 加密密钥存储在 `/data/.encryption_key`
- 管理员密码存储在 `/data/.admin_pass`
- 传输缓存存储在 `/data/cache/`

## 📝 更新日志

### v1.0.0
- ✅ 初始版本发布
- ✅ 支持夸克网盘 ↔ 移动云盘互传
- ✅ 实现 SHA256 秒传检测
- ✅ 添加 Web 管理界面
- ✅ 实现密码加密存储
- ✅ Docker 容器化部署

## 🤝 贡献

欢迎贡献代码！请提交 PR 到主分支。

## 📄 许可证

MIT License

## 🔗 相关链接

- [OpenList 官方文档](https://github.com/ulixee/openlist)
- [夸克网盘](https://pan.quark.cn/)
- [移动云盘](https://pan.xunlei.com/)
