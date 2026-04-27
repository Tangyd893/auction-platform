# Auction Platform 优化报告

日期：2026-04-27

## 优化范围

本次优化优先处理影响运行正确性、安全性和仓库可维护性的部分，避免引入大规模架构改动。

| 优化部分 | 发现的问题 | 优化方案 | 预期效果 | 状态 |
| --- | --- | --- | --- | --- |
| HTTP JWT 认证 | 受保护 REST 接口未统一校验 JWT，且原逻辑会在未登录或 token 无效时默认使用用户 ID 1 | 新增 Gin 认证中间件，受保护接口必须提供有效 `Bearer` token；用户 ID 从已验证 claims 注入上下文 | 阻止匿名用户创建拍品、出价、查看个人订单等越权操作，消除默认用户带来的安全风险 | 已完成 |
| JWT 校验 | token 校验未限制签名算法 | 在 `ValidateToken` 中限制 HMAC 签名算法 | 降低异常签名算法绕过风险 | 已完成 |
| 前端登录态 | 前端 API 拦截器读取 `localStorage.token`，但登录态主要由 Zustand persist 存到 `auth-storage`，导致请求可能不带 token | 登录时同步写入 `localStorage.token`，请求时同时兼容 `token` 和 `auth-storage` | 登录后创建拍品、出价、订单等接口能稳定带上认证头 | 已完成 |
| REST 响应适配 | 前端将拍品详情和订单详情解析为 `{ item }` / `{ order }` 包装结构，但后端返回的是实体对象 | 修正 `itemApi.get`、`orderApi.get`、`orderApi.updateStatus` 的类型和解析逻辑 | 修复拍品详情页、订单详情/状态更新潜在的空数据问题 | 已完成 |
| 出价状态流转 | 新出价创建后再查询最高价会查到新出价本身，旧领先出价可能不会被标记为 `outbid` | 使用 `MarkItemBidsOutbid(itemID, newBidID)` 批量淘汰旧 `active/winning` 出价，再将新出价标记为 `winning` | 保证同一拍品最多只有一个 `winning` 出价，提升竞价状态一致性 | 已完成 |
| 前端依赖与仓库卫生 | 已废弃 gRPC-Web 依赖仍保留在 `package.json`，且 `frontend/node_modules` 下生成文件被 Git 跟踪 | 移除废弃依赖、删除已跟踪的 `node_modules` 生成文件，新增 `.gitignore` | 降低安装体积和维护噪音，避免把本地依赖目录继续提交到仓库 | 已完成 |
| 旧 API 客户端 | `frontend/src/lib/api.ts` 未被引用，且仍包含旧路由 | 删除未使用的旧客户端，只保留 `frontend/src/api/rest.ts` | 减少重复入口和后续误用风险 | 已完成 |
| 本地配置 | 后端配置端口为 `8080`，但 README、Vite 代理和架构文档均使用 `8082` | 将 `backend/config.yaml` HTTP 端口统一为 `8082` | 前端代理、文档和后端运行配置保持一致 | 已完成 |
| 开发脚本 | `make seed` 指向不存在的 `cmd/server/seed.go` | 修正为 `backend/cmd/seed/main.go` | 恢复种子数据脚本入口 | 已完成 |
| 文档准确性 | README 和架构文档中 PostgreSQL 版本、数据访问层、REST 路由、前端目录描述过期 | 同步为当前实现：PostgreSQL 15、`database/sql + lib/pq`、当前 REST 路由和前端结构 | 降低新人启动和排错成本 | 已完成 |

## 后续建议

| 建议项 | 优化方案 | 预期效果 | 优先级 |
| --- | --- | --- | --- |
| 出价事务一致性 | 将创建出价、更新拍品当前价、淘汰旧出价放入同一个数据库事务，并考虑行级锁或乐观锁 | 避免并发出价时出现价格和出价状态不一致 | 高 |
| 后端认证覆盖 | 为 gRPC 请求补齐 token 解析和 context 注入，替换当前 `getUserIDFromContext` 默认值 | 让 gRPC 与 REST 的权限模型一致 | 高 |
| 前端类型收敛 | 移除页面中的 `any`，将 API 响应类型贯穿到列表、详情、订单页面 | 提前发现字段名和响应结构变更 | 中 |
| 依赖锁文件 | 在前端补充并提交 `package-lock.json` 或明确使用 pnpm/yarn 锁文件 | 提升安装可复现性 | 中 |
| CI 验证 | 增加 GitHub Actions：后端 `go test ./...`，前端 `npm ci && npm run build` | 上传后自动发现构建和测试回归 | 中 |
| README 编码与启动流程 | 统一文档编码，补充从 Docker、seed 到前后端启动的完整命令 | 降低本地运行门槛 | 低 |

## 验证说明

已使用 `E:\develop\Golang\go1.25.9.windows-amd64\go\bin\go.exe` 完成后端验证：

```bash
cd backend && go test ./...
```

结果：通过。

前端构建未继续执行。当前前端依赖目录不完整，建议在验收环境安装依赖后运行：

```bash
cd frontend && npm install && npm run build
```
