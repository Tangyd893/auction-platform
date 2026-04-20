# Auction Platform - 在线拍卖系统

Go + gRPC 实战项目：完整的 Clean Architecture 分层、gRPC 四种 RPC 模式、工程化测试与监控。

> ⚠️ **前端通信方案变更（2026-04-20）**：由于 gRPC-Web 与 Vite 5.x 存在系统性兼容冲突，前端从 gRPC-Web 切换为 REST API（HTTP/JSON）。后端 gRPC 服务保留，前端通过 Gin HTTP 接口通信。

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.21 + grpc-go v1.62 |
| HTTP | Gin v1.9.1 |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 前端 | React 18 + Vite |
| 代理 | Envoy 3（gRPC-Web 场景）|
| 监控 | Prometheus |
| ORM | GORM v2 |

## 架构

```
浏览器 ──HTTP/JSON──► Go HTTP :8082 (Gin REST API)
                          │
                          └── gRPC :50051（内部调用，后端直连）
                                   │
                         ┌─────────┴─────────┐
                         ▼                   ▼
                   PostgreSQL            Redis
                   (主数据)           (会话/限流/Pub/Sub)
```

> **历史方案（已废弃）**：浏览器 → gRPC-Web → Envoy → gRPC。此方案因 gRPC-Web 代码生成工具（protoc-gen-grpc-web v1.5.0）与 Vite 5.x 不兼容而废弃。详见 [docs/grpc-web-white-screen-bug.md](docs/grpc-web-white-screen-bug.md)。

## 快速启动

```bash
# 1. Docker 基础设施
cd docker && docker compose up -d

# 2. 后端（gRPC :50051 + HTTP :8082）
cd backend && go run cmd/server/main.go

# 3. 前端（Vite :3001）
cd frontend && npm run dev
```

访问 http://localhost:3001 ，管理员账号：`admin` / `admin123`

## REST API 端点

所有端点基准路径：`http://localhost:8082/api`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/auth/register | 注册用户 | — |
| POST | /api/auth/login | 登录，返回 JWT | — |
| POST | /api/items | 创建拍品 | JWT |
| GET | /api/items | 列出拍品（?status=&keyword=）| — |
| GET | /api/items/:id | 拍品详情 | — |
| GET | /api/items/my | 我的拍品 | JWT |
| DELETE | /api/items/:id | 取消拍品 | JWT |
| POST | /api/bids | 出价（?itemId=&amount=）| JWT |
| GET | /api/bids/:itemId | 拍品出价记录 | — |
| GET | /api/bids/my | 我的出价 | JWT |
| POST | /api/orders | 创建订单 | JWT |
| GET | /api/orders | 我的订单列表 | JWT |
| GET | /api/orders/:id | 订单详情 | JWT |
| PUT | /api/orders/:id/status | 更新订单状态 | JWT |

> **StreamBids 实时竞价**：Server Streaming via gRPC 已废弃，降级为轮询方案（TBD）。

## 项目结构

```
auction-platform/
├── docs/
│   ├── 项目概述与技术架构.md
│   └── grpc-web-white-screen-bug.md   ← gRPC-Web 兼容问题完整记录
├── proto/
│   └── auction.proto                   ← gRPC Protobuf 定义（后端保留）
├── backend/
│   ├── cmd/server/                    ← 入口：gRPC :50051 + HTTP :8082
│   └── internal/
│       ├── config/                    ← YAML 配置
│       ├── handler/                   ← HTTP Handler（Gin）+ gRPC Handler
│       ├── interceptor/               ← JWT 拦截器
│       ├── model/                    ← 数据模型
│       ├── repository/               ← 数据访问层
│       └── service/                  ← 业务逻辑 + 单元测试
├── frontend/
│   └── src/
│       ├── api/                      ← REST API 客户端（TODO: 重写）
│       ├── grpc/                     ← gRPC-Web 代码（待删除）
│       └── pages/                    ← React 页面组件
└── docker/
    └── docker-compose.yml
```

## gRPC RPC 模式（后端全部实现）

| 模式 | 用途 | 状态 |
|------|------|------|
| Unary | Login, GetItem, PlaceBid | ✅ gRPC 已实现 |
| Client Streaming | PlaceBidBatch | ✅ gRPC 已实现 |
| Server Streaming | StreamBids | ✅ gRPC 已实现（前端降级为轮询）|
| Bidirectional | BidirectionalBid | ✅ gRPC 已实现（前端通过 REST 降级）|

## 测试

```bash
# 后端单元测试（无需数据库）
cd backend && go test -v ./internal/service/...

# E2E 测试（需前后端运行）
node test-e2e.mjs
```

## 监控

Prometheus 指标：`http://localhost:9090`

- `bids_total{result}` — 出价总数（success/rejected/error）
- `bid_latency_seconds` — 出价处理延迟
- `circuit_breaker_state` — 熔断器状态

## 工程化安全

- **限流**：per-user 每秒最多 5 次出价（滑动窗口）
- **熔断器**：连续 20 次失败自动熔断 30 秒
- **Nonce 防重放**：`PlaceBidRequest.nonce + timestamp`
- **JWT 拦截器**：所有 gRPC/HTTP 调用需带有效 Token

## 文档

- [项目概述与技术架构](./docs/项目概述与技术架构.md)
- [gRPC-Web 白屏问题调试记录](./docs/grpc-web-white-screen-bug.md)
- [Go 项目开发测试流程](../forRead/Go项目开发测试流程.md)
