## Context

仓库已有 `stock-center`（8081）、`stock-manage`（8082）、`stock-front`（Vue 3 + Vite）骨架。当前 `Stock{SKU, Warehouse, Quantity int64}` 与单次 `inbound` 示例不足以表达不锈钢业务。

约束（来自已批准设计 spec）：

- 双服务 + 本机 MySQL 8（`stock_center`、`stock_manage` 两库）
- 物料档案在 center；单据在 manage；前端只访问 manage
- 原材料 kg、成品件/米；销售单驱动出库；加工损耗 = 投入 − 产出
- 第一版单人无登录；表预留 `orgId`、`createdBy`

## Goals / Non-Goals

**Goals:**

- 实现物料/批次/双计量库存/流水与四类单据全流程
- manage 确认单据后编排调用 center，保证库存真源唯一
- 启动器一键拉起双服务 + 浏览器；manage 内嵌前端
- 列表筛选 API 使用 `qp-<field>-<operator>` 谓词参数

**Non-Goals:**

- RBAC、多租户、财务模块、边角料分类库存
- 服务间 JWT（第一版 center 仅本机可信调用）
- Docker 云部署（文档预留即可）

## Decisions

### D1：双服务保留，不合并单体

**选择**：继续 `stock-center` + `stock-manage`。  
**理由**：业务膨胀后权限、多业务线放在 manage；库存真源隔离在 center。  
**备选**：单体 `stock-app` — 打包更简单但违背用户扩展诉求。

### D2：金额与重量用 `decimal.Decimal`

**选择**：GORM + `shopspring/decimal`，JSON 传字符串。  
**理由**：避免 float 精度问题（重量、单价）。  
**备选**：`int64` 存厘克 — 前端不友好。

### D3：旧 `stock` 表弃用策略

**选择**：新表 `material`、`material_batch`、`stock_balance`、`stock_ledger` 并行；删除旧路由；无生产数据，直接 AutoMigrate 新表。  
**理由**：骨架阶段无迁移负担。  
**备选**：保留旧 API 兼容 — YAGNI。

### D4：单号生成

**选择**：manage 内 `doc_no` 生成器，格式 `{PREFIX}{YYYYMMDD}{4位序号}`（如 `IN202607030001`），按日重置序号存在 `doc_sequence` 表。  
**理由**：单人本地够用，不引入 Redis。

### D5：前端嵌入 manage

**选择**：`//go:embed dist/*`，Gin `NoRoute` 回退 `index.html`；开发时仍 `pnpm dev` 代理到 manage。  
**理由**：分发单一二进制体验。  
**备选**：独立 nginx — 本地部署过重。

### D6：manage 作为 BFF

**选择**：物料/库存只读与写操作均由 manage handler 转发 center；前端不持有 center 地址。  
**理由**：统一 CORS、未来鉴权关口。

### D7：加工单确认事务

**选择**：manage 在 DB 事务内更新单据状态；center 调用顺序：先全部 outbound 再全部 inbound；任一 center 失败则回滚单据状态并返回错误（第一版不做分布式事务，依赖操作员重试）。  
**理由**：两库两服务，避免 2PC 复杂度。  
**缓解**：确认前 manage 可先调 center 查询余额校验。

## Risks / Trade-offs

- [双进程启动失败] → 启动器健康检查与明确日志路径
- [manage 确认后 center 部分失败] → 确认前余额校验；错误时不改 `confirmed`；操作日志记失败
- [旧代码引用 SKU API] → 本变更内删除旧 handler 与 client 方法，编译期发现断裂
- [双库备份遗漏] → README 提供 `mysqldump` 两库脚本

## Migration Plan

1. 开发环境：本机 MySQL 创建两库，`.env` 配置
2. 部署新表 AutoMigrate，不保留旧 `stock` 数据
3. 替换 manage `inventory_service` 为新单据服务
4. 前端切换到新路由；移除 wujie 相关（第一版独立运行，非微前端子应用）
5. 打包：`make build-all` 产出二进制 + 启动脚本

**Rollback**：恢复上一 git tag；数据库备份还原。

## Open Questions

- （无）设计 spec 已用户确认。
