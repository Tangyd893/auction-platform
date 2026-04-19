# Auction Platform - 在线拍卖系统

Go + gRPC 实战项目：在线拍卖系统。展示完整的 Clean Architecture 分层、gRPC 四种 RPC 模式、gRPC-Web 前端集成、工程化测试与监控。

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.21 + grpc-go v1.62 |
| HTTP | Gin v1.9.1 |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 前端 | React 18 + Vite |
| gRPC-Web | @improbable-eng/grpc-web 0.15.0 |
| 代理 | Envoy 3 |
| 监控 | Prometheus |
| ORM | GORM v2 |
| E2E 测试 | Playwright |

## 架构

```
Browser ──gRPC-Web──► Envoy :8081 ──HTTP/2──► Go gRPC :50051
Browser ──HTTP─────► Go HTTP  :8080 (REST API)

Go Backend
  ├── Handler    (gRPC + HTTP/REST)
  ├── Service   (业务逻辑，接口抽象)
  └── Repository (PostgreSQL + Redis)
```

## 快速启动

```bash
# 1. Docker 基础设施
cd docker && docker compose up -d

# 2. 后端
cd backend && go run cmd/server/main.go

# 3. 前端（另一个终端）
cd frontend && npm run dev
```

访问 http://localhost:3001 ，管理员账号：`admin` / `admin123`

## gRPC RPC 模式（全部实现）

| 模式 | 用途 | 状态 |
|------|------|------|
| Unary | Login, GetItem, PlaceBid | ✅ |
| Client Streaming | PlaceBidBatch（批量出价）| ✅ |
| Server Streaming | StreamBids（实时订阅拍品出价）| ✅ |
| Bidirectional | BidirectionalBid（一个连接同时出价+订阅）| ✅ 服务端完整 |

> **Bidirectional Streaming 前端说明**：gRPC-Web 0.15.0 浏览器版本不支持 bidirectional streaming。服务端已完整实现，前端通过 **Server Streaming (StreamBids)** 实现实时竞价。StreamAuction（拍卖大厅）也支持多 item 批量订阅。

## 项目结构

```
backend/
├── cmd/server/          # 入口：gRPC :50051 + HTTP :8080 + Prometheus :9090
├── internal/
│   ├── config/          # YAML 配置
│   ├── handler/         # HTTP Handler
│   ├── interceptor/     # JWT 拦截器
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   └── service/        # 业务逻辑 + 单元测试
│       ├── auction_service.go   # gRPC 所有 RPC 实现
│       ├── bid_service.go       # 出价核心逻辑
│       ├── ratelimit.go         # 滑动窗口限流
│       ├── circuitbreaker.go    # 熔断器
│       └── bid_service_test.go  # 8 个单元测试
frontend/
├── src/grpc/
│   ├── client.ts        # gRPC-Web API 封装
│   └── proto/           # 生成的 JS gRPC 代码
├── pages/               # React 页面组件（9个）
└── tests/              # Playwright E2E 测试
```

## 测试

```bash
# 后端单元测试（无需数据库）
cd backend && go test -v ./internal/service/...

# E2E 测试（需要前后端运行）
node test-e2e.mjs
# 或使用 Playwright（需 npm install）：
cd frontend && npm run test:e2e
```

**E2E 测试覆盖**：登录、页面导航、创建拍品、拍品列表、详情页、出价、我的出价、StreamBids 连接、控制台错误检查

## 监控

Prometheus 指标：`http://localhost:9090`

- `bids_total{result}` — 出价总数（success/rejected/error）
- `bid_latency_seconds` — 出价处理延迟
- `bidirectional_streams_active` — 活跃双向流数
- `circuit_breaker_state` — 熔断器状态

## 工程化安全

- **限流**：per-user 每秒最多 5 次出价（滑动窗口）
- **熔断器**：连续 20 次失败自动熔断 30 秒
- **连接数上限**：全局最多 500 个双向流
- **Nonce 防重放**：`PlaceBidRequest.nonce + timestamp`
- **JWT 拦截器**：所有 gRPC 调用需带有效 Token
- **Prometheus 监控**：5 个核心指标

## 文档

- [项目概述与技术架构](./docs/项目概述与技术架构.md)
- [Go 项目开发测试流程](../forRead/Go项目开发测试流程.md)

## 支付集成说明

支付模块**未集成**（按要求留白）。接口已预留（`CreatePayment`、`ExecutePayment`），实际支付需接入支付宝/微信/Stripe 等。
