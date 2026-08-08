## Context

`sales-front` 已用 Dexie 落地销售单 + 开票（库名 `sales-front`：表 `salesOrders`、`invoiceDocs`）。页面只依赖 `*Api`；开票创建/作废在前端用跨表事务。同仓 `stock-manage` 另有库存域 `/api/v1/sales-orders`（物料/批次/确认/发货），与本工具模型不同，不可复用。

kb 约定：`domains/personal-go/sales-manage-api.md`、`domains/workflow/local-first-to-go-mysql.md`。

## Goals / Non-Goals

**Goals:**

- 独立服务 `sales-manage` 成为销售开票唯一写源（MySQL）
- HTTP JSON 对齐现有前端 `SalesOrder` / `InvoiceDoc`（嵌套 `lines`、字符串 UUID id）
- 列表与开票候选筛选后端化（`qp-*`）
- 开票创建/作废单请求事务，语义对齐现 `invoiceDocApi`
- 前端保留 Api 函数名与页面结构，内部换 `fetch`

**Non-Goals:**

- IndexedDB → MySQL 数据导入脚本（第二期）
- 与 stock-* 物料/库存/发货打通
- 登录 / 多租户 / 权限
- 打印样式改造
- 长期 IndexedDB 离线双写

## Decisions

### 1. 独立服务，不用 stock-manage

- **选择**：`apps/sales-manage` + 库 `sales_manage` + 端口 `8083`
- **理由**：模型、生命周期、开票能力与库存销售单不同；混路由易踩踏
- **备选**：挂 stock-manage 旁路新路由 → 拒绝（同进程同名风险高）

### 2. URL：`/api/v1/sale/...`

| 资源 | 路径 |
|------|------|
| 销售单 | `/api/v1/sale/orders` |
| 开票候选 | `GET /api/v1/sale/orders/invoice-candidates`（须注册在 `/:id` 之前） |
| 开票单 | `/api/v1/sale/invoice-docs` |
| 作废 | `POST /api/v1/sale/invoice-docs/:id/void` |

- **理由**：用户要求 `/sale/orders` 风格；补 `/api/v1` 与同仓一致；单数 `sale` 区别库存 `sales-orders`
- **备选**：扁平 `/api/v1/sales-orders` → 易与库存混淆

### 3. 表结构：头行拆分，响应仍嵌套

- `sales_order` / `sales_order_line`（`invoice_doc_id` 可空）
- `invoice_doc` / `invoice_doc_line`（含来源销售单快照字段）
- 对外 id：字符串 UUID（与现前端一致）
- 金额/数量：`decimal` 入库，JSON 可与现 number 对齐（实现时与前端联调一致即可）

### 4. `qp-*` 白名单（首版）

销售单列表：`customerName` like；`orderNo` eq / in / like；`orderDate` gte / lte；`warehouseName` eq（可选）。  
开票列表：`status` eq；`invoiceNo` like。  
候选：复用销售单筛字段 + 固定条件 `needInvoice=true AND invoice_doc_id IS NULL`。  
分页：`page`、`pageSize`；默认排序 `updated_at DESC`。  
非法 field/operator → 400。实现参考 `stock-center/internal/pkg/qp`。

### 5. 删除销售单策略

- **选择**：若任一明细存在非空 `invoiceDocId`，拒绝删除（明确错误文案）
- **理由**：避免开票单悬空；现前端「不级联」语义过弱
- **备选**：允许删除 → 二期再收紧

### 6. 前端切换

- 保留 `listSalesOrders` 等导出；内部 HTTP；列表筛选项映射为 `qp-*`
- `createInvoiceDoc` / `voidInvoiceDoc` 各打一枪，不再本地事务
- 删除 Dexie 写路径与依赖；单测改为 mock HTTP 或测纯函数
- Vite 代理 / env：`VITE_SALES_API_BASE` → `http://localhost:8083`

### 7. 分层

与 stock-manage 同构：`handler → service → repository → model`；开票事务在 service；统一 `pkg/response`。

## Risks / Trade-offs

- [空库直切丢本地数据] → 文档标明；第二期导入；用户可先自备导出
- [与库存销售单命名混淆] → 路由 `/sale/*`、独立服务与文档反复声明
- [Gin 路由 `invoice-candidates` 被 `:id` 吃掉] → 静态路径先注册
- [列表从全量改分页] → 前端列表接 `total`；筛选项必须带进 query
- [decimal vs JS number] → 首版金额数量用可 JSON 数字；边界用例单测覆盖

## Migration Plan

1. 实现并本地跑通 `sales-manage` + 空库
2. 前端切 HTTP，确认 CRUD / 开票 / 作废
3. 停用 IndexedDB；更新 README 与仓内/kb 文档
4. 回滚：恢复前端 Dexie 分支（无服务端数据需保留时）；服务端可停进程
5. 第二期：IndexedDB JSON 导出 + 导入接口/脚本

## Open Questions

- （无阻塞）是否在首版加 CORS 仅开发用 / 同域反向代理：实现时按 Vite proxy 优先
- 开票候选路径挂在 `orders` 下已定；若前端更偏好 `invoice-docs/candidates` 可在实现前微调，不影响表设计
