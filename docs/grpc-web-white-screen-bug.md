# gRPC-Web 白屏问题调试文档

## 问题描述

Firefox 149 下访问前端 `http://localhost:3001` 白屏，控制台报错：

```
[grpc-web-shim] grpc.web.* API ready
Uncaught TypeError: goog.exportSymbol is not a function   ← 第1个错误
Uncaught TypeError: can't access property "AuctionServiceClient", proto.auction is undefined  ← 第2个错误
Uncaught TypeError: can't access property "AuctionServiceClient", svc is undefined  ← 第3个错误
```

## 技术背景

### 参与组件

| 组件 | 职责 | 版本 |
|------|------|------|
| `google-protobuf` | protobuf JS 运行时 | npm 版本 |
| `@improbable-eng/grpc-web` | gRPC-Web 客户端 | 0.15.0 |
| `protoc-gen-grpc-web` | proto 代码生成器 | v1.5.0 |
| Vite | 前端构建工具 | 5.x |
| React | 前端框架 | 18 |

### 生成的 proto 代码结构

`protoc-gen-grpc-web v1.5.0` 生成的 JS 文件依赖 **Closure Library** 风格：

```javascript
// auction_pb.js（消息定义）
var jspb = require('google-protobuf');    // 引用 CommonJS 模块
var goog = jspb;                         // goog === jspb
goog.exportSymbol('proto.auction.Bid', null, global);  // 导出到全局

// auction_grpc_web_pb.js（服务客户端）
const grpc = {};
grpc.web = require('grpc-web');           // 引用 CommonJS 模块
const proto = {};
proto.auction = require('./auction_pb.js'); // 引用 CommonJS 模块
proto.auction.AuctionServiceClient = function(...) { ... };
```

## 根因分析

### 问题链路（5层）

```
┌─────────────────────────────────────────────────────────────────────┐
│ 第1层：Vite dev 模式不转换 CommonJS require()                       │
│   src/grpc/proto/*.js 是 CommonJS 格式，含 require('google-protobuf')│
│   Vite 直接作为 ESM 发送，浏览器遇到 require 报错                    │
└─────────────────────────────┬───────────────────────────────────────┘
                              │  [尝试方案] 把 proto 文件放入 node_modules
                              │  → Vite optimizeDeps 预Bundle，转换 require → ESM
                              │  → 但仍报错 ↓
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 第2层：google-protobuf UMD 不导出 goog.exportSymbol                   │
│   google-protobuf 的 UMD build：                                    │
│     "object" === typeof exports  → 走 CommonJS 分支                  │
│     不走 else (浏览器) 分支 → 从不执行 g("jspb.Map",...) → 不设      │
│     window.jspb.exportSymbol                                         │
│   实际 g() 函数写的是 window.jspb.Map = ...（小写）                 │
│   生成的 proto 文件：var goog = jspb; goog.exportSymbol() 失败       │
└─────────────────────────────┬───────────────────────────────────────┘
                              │  [尝试方案] 在 index.html <script> 标签
                              │  手动注入 exportSymbol/inherits 补丁
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 第3层：google-protobuf 每次 g() 调用会覆盖 window.jspb               │
│   google-protobuf.js 内: g("jspb.Map", r)                            │
│   g() 函数体: window.jspb = {}（每次调用都覆盖）                    │
│   如果 exportSymbol 在 google-protobuf 加载前注入，会被清空          │
│   如果在加载后注入，顺序正确                                          │
└─────────────────────────────┬───────────────────────────────────────┘
                              │  [尝试方案] 在 google-protobuf 加载后
                              │  立即注入 exportSymbol/inherits
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 第4层：protoc-gen-grpc-web v1.5.0 生成代码依赖新版 gRPC-Web API       │
│   生成代码使用了 grpc.web.MethodType.UNARY                          │
│   @improbable-eng/grpc-web 0.15.0 UMD build 不包含 MethodType       │
│   → grpc.web.MethodType is undefined                                 │
└─────────────────────────────┬───────────────────────────────────────┘
                              │  [尝试方案] 手动补充 MethodType 等 shim
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 第5层：proto 文件 require() 被 Vite 转换后，grpc-web 模块名失效      │
│   require('grpc-web') → 转换后指向 window.grpc                       │
│   但 proto 文件是 window.proto.auction 别名套别名                   │
│   加载顺序及时机导致 proto.auction 为 undefined                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 关键发现

**google-protobuf UMD 的 module.exports 不暴露 exportSymbol：**

```javascript
// google-protobuf.js 末尾
"object"===typeof exports  // Node.js: true → exports.exportSymbol = ma
// 浏览器: false → 走 else 分支，不设 window.jspb.exportSymbol
```

**google-protobuf 用小写 `jspb` 设全局，不是 `JSPB`：**

```javascript
// g() 函数内部：g("jspb.Map", r, void 0)
window.jspb = window.jspb || {};  // 小写
window.jspb.Map = r;             // 小写
```

## 尝试过的修复方案

### 方案 A：Closure Library Shim（推荐，失败）

```typescript
// main.tsx 顶部
import 'google-closure-library/closure/goog/bootstrap/webworkers';
```

**结论**：google-closure-library 也没有直接导出 `goog.exportSymbol`，且与 Vite 的 ESM 处理存在冲突。

### 方案 B：手动注入 exportSymbol/inherits（部分成功）

在 `index.html` 中，google-protobuf 加载后立即注入：

```html
<script src="/node_modules/google-protobuf/google-protobuf.js"></script>
<script>
  window.goog = window.jspb;  // 关键：google-protobuf 设的是 window.jspb
  window.goog.exportSymbol = function(path, symbol) {
    var parts = path.split('.'), t = window;
    for (var i = 0; i < parts.length - 1; i++)
      t = t[parts[i]] || (t[parts[i]] = {});
    t[parts[parts.length - 1]] = symbol;
  };
  window.goog.inherits = function(ctor, superCtor) { /* 标准 inherits */ };
</script>
```

**进展**：解决了 `goog.exportSymbol is not a function`，但遇到 `grpc.web.MethodType is undefined`。

### 方案 C：手动注入 grpc.web shim（部分成功）

```javascript
window.grpc.web.MethodType = { UNARY: 0, SERVER_STREAMING: 1, ... };
window.grpc.web.MethodDescriptor = function(...) { ... };
window.grpc.web.GrpcWebClientBase = function(...) { ... };
```

**进展**：解决了 `MethodType is undefined`，但遇到 `proto.auction is undefined`（grpc-web 模块名冲突）。

## 最终结论

**gRPC-Web 与 Vite 5.x 存在系统性兼容冲突**，不是因为某个特定 bug，而是因为：

1. **生成代码格式**：protoc-gen-grpc-web v1.5.0 生成的 CommonJS/Closure 风格代码，与 Vite 的 ESM 处理不兼容
2. **依赖版本断层**：生成代码期望新版 gRPC-Web API，但 @improbable-eng/grpc-web 0.15.0 缺少关键类型
3. **多层兼容补丁脆弱**：任何一层版本变化都会导致整体失效

## 正确解决方案

### 方案：切换为 REST API（推荐）

Go 后端已有完整的 Gin HTTP/REST 接口，前端直接用 `fetch` 调用，无需任何 proto 依赖：

```
浏览器 ──HTTP/JSON──► Go HTTP :8082 (Gin)  ← 原生支持，无代理
                              │
                              └── gRPC :50051 (内部调用)
```

**优势**：
- 浏览器原生支持 HTTP/JSON，无需 gRPC-Web 代理
- 调试工具丰富（DevTools Network 面板直接看 JSON）
- 无 Closure Library 依赖冲突
- 前后端完全解耦，proto 文件变更不影响前端

**劣势**：
- 失去 gRPC 的类型安全（需手动维护接口契约）
- Server Streaming 需降级为轮询或 WebSocket
- 带宽效率低于 protobuf 二进制

### 迁移步骤

1. 前端 `src/grpc/client.ts` → `src/api/rest.ts`，用 `fetch` 调 Gin HTTP 接口
2. 移除 `index.html` 中的所有 `<script>` 标签（proto/grpc-web 相关）
3. 删除 `src/grpc/proto/` 目录
4. 删除 `src/grpc/grpc-web-shim.js`
5. 更新 `vite.config.ts` 移除 grpc-web 相关配置
6. Server Streaming 降级为轮询（`setInterval` + REST `GetBids`）

### 保留 gRPC 的场景

如果未来需要 gRPC：
- 使用 `ts-proto` 生成纯 ES6/TypeScript 代码（不依赖 Closure Library）
- 使用 `connectrpc` 代替 `protoc-gen-grpc-web`（生成浏览器原生 fetch 代码）

## 相关文件

- `frontend/index.html` — gRPC-Web script 注入（待删除）
- `frontend/src/grpc/client.ts` — gRPC-Web 封装（待迁移）
- `frontend/src/grpc/proto/auction_pb.js` — 生成的消息代码（待删除）
- `frontend/src/grpc/proto/auction_grpc_web_pb.js` — 生成的服务代码（待删除）
- `frontend/src/grpc/grpc-web-shim.js` — gRPC-Web shim（待删除）
- `backend/internal/handler/http.go` — Gin HTTP 接口（REST 基础）
- `proto/auction.proto` — Protobuf 定义（后端 gRPC 保留）
