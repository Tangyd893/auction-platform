# Auction Platform - 在线拍卖系统

基于 Go + gRPC + Gin + React 18 的实时拍卖平台。

## 技术栈

### 后端 (Go)

| 层级 | 技术 |
|------|------|
| RPC 框架 | google.golang.org/grpc v1.62 |
| HTTP 框架 | github.com/gin-gonic/gin |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 数据层 | github.com/jmoiron/sqlx |
| 配置 | github.com/spf13/viper |
| 日志 | github.com/rs/zerolog |
| 指标 | github.com/prometheus/client_golang |
| 认证 | github.com/golang-jwt/jwt/v5 |

### 前端 (React)

| 层级 | 技术 |
|------|------|
| 框架 | React 18 + Vite + TypeScript |
| 状态管理 | Zustand + TanStack Query |
| UI | Tailwind CSS |
| 图表 | recharts |

### Proto 服务设计

```
AuctionService (gRPC)
├── Register/Login             (Unary)          用户注册/登录
├── CreateItem                 (Unary)          创建拍品
├── GetItem/ListItems          (Unary)          查询拍品
├── PlaceBid                   (Unary)          出价
├── PlaceBidBatch              (Client Streaming) 批量出价
├── StreamBids                 (Server Streaming) 订阅出价更新
├── BidirectionalBid           (Bidirectional)    实时竞价窗口
├── StreamAuction              (Server Streaming) 拍卖大厅
└── CreateOrder/GetOrder       (Unary)          订单管理
```

## 快速开始

### 前置依赖

```bash
# 安装 protoc (Linux)
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
sudo unzip -o protoc-25.1-linux-x86_64.zip -d /usr/local

# 安装 Go protobuf 插件
export PATH=$PATH:/usr/local/go/bin:~/go/bin
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
```

### 1. 启动基础设施

```bash
cd docker
docker compose up -d
```

### 2. 初始化数据库

```bash
# 创建默认管理员账号
cd ../backend
go run cmd/seed/main.go
```

### 3. 启动后端

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

服务端口：
- gRPC: `localhost:50051`
- HTTP: `localhost:8080`
- Prometheus: `localhost:9090`

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

访问 `http://localhost:3001`

### 默认账号

```
用户名: admin
密码:   admin123
```

## 项目结构

```
auction-platform/
├── proto/
│   ├── auction.proto            # Proto 定义文件
│   └── gen/auction/            # 生成的 Go 代码
├── backend/
│   ├── cmd/
│   │   ├── server/             # 主服务入口
│   │   └── seed/               # 数据库初始化
│   ├── internal/
│   │   ├── config/             # Viper 配置
│   │   ├── model/              # 数据模型
│   │   ├── repository/          # PostgreSQL + Redis 操作
│   │   ├── service/            # 业务逻辑
│   │   └── interceptor/         # gRPC 拦截器
│   ├── proto/gen/auction/      # 生成的 proto 代码
│   └── config.yaml
├── frontend/
│   └── src/
│       ├── pages/              # 页面组件
│       ├── components/          # 公共组件
│       ├── stores/             # Zustand 状态
│       └── lib/                # API 封装
├── docker/                     # Docker Compose
└── Makefile
```

## Makefile 命令

```bash
make docker-start    # 启动 Docker 基础设施
make docker-stop     # 停止 Docker
make proto           # 生成 proto 代码
make backend-deps    # 下载 Go 依赖
make backend-run     # 启动后端
make seed            # 初始化管理员账号
make frontend-deps   # 安装前端依赖
make frontend-run    # 启动前端
make dev             # 一键启动（基础设施+后端+前端）
```

## gRPC 四种 RPC 模式

| 模式 | 用途 | 实现 |
|------|------|------|
| Unary | 注册/登录/查询 | `Register`, `Login`, `GetItem`, `PlaceBid` |
| Client Streaming | 批量出价 | `PlaceBidBatch` |
| Server Streaming | 实时推送 | `StreamBids`, `StreamAuction` |
| Bidirectional | 实时竞价 | `BidirectionalBid` |

## 功能模块

- [x] 用户注册/登录 (JWT 认证)
- [x] 拍品发布/编辑/下架
- [x] 实时竞价
- [x] 出价历史记录
- [x] 成交订单管理
- [x] Prometheus 指标
- [ ] gRPC-Web 前端集成
- [ ] WebSocket 实时竞价
- [ ] 支付集成
