# Auction Platform - 在线拍卖系统

基于 Go + gRPC + Gin + React 18 的实时拍卖平台。

## 技术栈

### 后端 (Go)

| 层级 | 技术 |
|------|------|
| RPC 框架 | google.golang.org/grpc v1.62 |
| HTTP 框架 | github.com/gin-gonic/gin v1.9.1 |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 数据层 | github.com/jmoiron/sqlx |
| 配置 | github.com/spf13/viper |
| 日志 | github.com/rs/zerolog |
| 指标 | github.com/prometheus/client_golang |
| 认证 | github.com/golang-jwt/jwt/v5 |
| 测试 | github.com/stretchr/testify |

### 前端 (React)

| 层级 | 技术 |
|------|------|
| 框架 | React 18 + Vite + TypeScript |
| 状态管理 | Zustand + TanStack Query |
| 通信 | **gRPC-Web** + @improbable-eng/grpc-web |
| UI | Tailwind CSS |
| 图表 | recharts |

### 架构

```
Browser (gRPC-Web)
    │
    │  HTTP/2 (binary frames)
    ▼
┌─────────────────────┐
│  Envoy Proxy        │  ← 协议转换 (gRPC-Web → gRPC)
│  :8081              │
└────────┬────────────┘
         │  gRPC (Protobuf binary)
         ▼
┌─────────────────────┐
│  Go gRPC Server     │
│  :50051             │
└─────────────────────┘

Browser (REST/HTTP)   ← 备选路径（开发调试）
    │
    ▼
┌─────────────────────┐
│  Go Gin HTTP Server │
│  :8080              │
└─────────────────────┘
```

### Proto 服务设计

```
AuctionService (gRPC)
├── Register/Login             (Unary)            用户注册/登录
├── CreateItem                 (Unary)            创建拍品
├── GetItem/ListItems          (Unary)            查询拍品
├── PlaceBid                   (Unary)            出价
├── PlaceBidBatch              (Client Streaming)  批量出价
├── StreamBids                 (Server Streaming)  订阅出价更新 ✅ 实时
├── BidirectionalBid            (Bidirectional)     实时竞价窗口
├── StreamAuction               (Server Streaming) 拍卖大厅
└── CreateOrder/GetOrder      (Unary)            订单管理
```

## 快速开始

### 前置依赖

```bash
# Go 1.21+
go version

# protoc 25.1+
protoc --version

# Go protobuf 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
go install github.com/grpc/grpc-web/cmd/protoc-gen-grpc-web@latest

# JS protobuf 插件
npm install -g protoc-gen-js
```

### 1. 启动基础设施

```bash
cd docker
docker compose up -d
```

### 2. 生成代码

```bash
# Proto Go 代码（Go 后端）
cd backend
go mod tidy
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       ../proto/auction.proto

# Proto JS 代码（前端）
protoc --proto_path=. \
       --js_out=import_style=commonjs:frontend/src/grpc \
       --grpc-web_out=import_style=commonjs,mode=grpcwebtext:frontend/src/grpc \
       proto/auction.proto
```

### 3. 初始化数据库

```bash
# 创建默认管理员账号（用户名: admin / 密码: admin123）
go run cmd/seed/main.go
```

### 4. 启动后端

```bash
cd backend
go run cmd/server/main.go
```

服务端口：
- gRPC: `localhost:50051`
- HTTP: `localhost:8080`
- Prometheus: `localhost:9090`

### 5. 启动前端

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
│   ├── auction.proto              # Proto 定义文件
│   └── gen/auction/             # 生成的 Go 代码
├── backend/
│   ├── cmd/
│   │   ├── server/              # 主服务入口
│   │   └── seed/               # 数据库初始化
│   ├── internal/
│   │   ├── config/             # Viper 配置加载
│   │   ├── handler/            # HTTP Handler (Gin)
│   │   ├── interceptor/        # gRPC 拦截器
│   │   ├── model/              # 数据模型
│   │   ├── repository/         # PostgreSQL + Redis 操作
│   │   └── service/           # 业务逻辑层
│   │       └── *_test.go       # 单元测试
│   └── config.yaml
├── frontend/
│   └── src/
│       ├── grpc/               # gRPC-Web 客户端
│       │   ├── client.ts       # API 封装
│       │   └── proto/          # 生成的 JS 代码
│       ├── pages/              # 页面组件
│       ├── components/         # 公共组件
│       ├── stores/             # Zustand 状态
│       └── lib/                # 工具库
├── docker/
│   ├── docker-compose.yml
│   └── envoy.yaml             # gRPC-Web 代理配置
└── Makefile
```

## Makefile 命令

```bash
make docker-start    # 启动 Docker 基础设施
make docker-stop    # 停止 Docker
make proto         # 生成 proto 代码（Go + JS）
make backend-deps  # 下载 Go 依赖
make backend-run   # 启动后端
make seed          # 初始化管理员账号
make frontend-deps # 安装前端依赖
make frontend-run  # 启动前端
make dev           # 一键启动
make test          # 运行后端测试
make build         # 编译后端
```

## 测试

### 后端测试（Go）

```bash
go test -v ./...              # 所有测试
go test -v -cover ./...      # 带覆盖率
go test -race -v ./...       # 竞态检测
```

当前测试状态：**8/8 通过**

```
=== RUN   TestPlaceBid_Success
--- PASS: TestPlaceBid_Success (0.00s)
=== RUN   TestPlaceBid_TooLow
--- PASS: TestPlaceBid_TooLow (0.00s)
=== RUN   TestPlaceBid_CannotBidOwnItem
--- PASS: TestPlaceBid_CannotBidOwnItem (0.00s)
=== RUN   TestPlaceBid_AuctionEnded
--- PASS: TestPlaceBid_AuctionEnded (0.00s)
=== RUN   TestPlaceBid_ItemNotActive
--- PASS: TestPlaceBid_ItemNotActive (0.00s)
=== RUN   TestCreateItem_EndTimeInPast
--- PASS: TestCreateItem_EndTimeInPast (0.00s)
=== RUN   TestCreateItem_StartAfterEnd
--- PASS: TestCreateItem_StartAfterEnd (0.00s)
=== RUN   TestCancelItem_AlreadySold
--- PASS: TestCancelItem_AlreadySold (0.00s)
PASS
```

### gRPC 调试

```bash
# 启动 reflection
grpcurl -plaintext localhost:50051 list

# 调用 Login
grpcurl -plaintext -d '{"username":"admin","password":"admin123"}' \
  localhost:50051 auction.AuctionService/Login
```

## gRPC 四种 RPC 模式

| 模式 | 用途 | 实现 |
|------|------|------|
| Unary | 注册/登录/查询 | `Register`, `Login`, `GetItem`, `PlaceBid` |
| Client Streaming | 批量出价 | `PlaceBidBatch` |
| Server Streaming | 实时推送 | `StreamBids`（前端已集成✅）|
| Bidirectional | 实时竞价 | `BidirectionalBid` |

## 功能模块

- [x] 用户注册/登录 (JWT 认证)
- [x] 拍品发布/编辑/下架
- [x] **gRPC-Web 前端集成**（全双工实时通信）
- [x] **Server Streaming 出价订阅**（`StreamBids`）
- [x] 实时竞价
- [x] 出价历史记录
- [x] 成交订单管理
- [x] Prometheus 指标
- [x] gRPC 拦截器（日志 + metrics）
- [x] 单元测试（Service 层，8/8 通过）
- [ ] Bidirectional Streaming 实时竞价（BidirectionalBid）
- [ ] StreamAuction 拍卖大厅
- [ ] 支付集成

## 项目完成度

| 模块 | 完成度 |
|------|--------|
| Proto 定义 | 100% |
| 数据库层 | 100% |
| Repository 层 | 100% |
| Service 层 | 100% |
| gRPC Server | 100% |
| HTTP Handler | 100% |
| 单元测试 | 100% |
| **gRPC-Web 前端集成** | **100%** |
| 前端页面 | 100% |
| Envoy 代理 | 100% |
