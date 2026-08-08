## Why

`sales-front` 以浏览器 IndexedDB 为真源，无法多机共用、备份与后续对接库存。领域模型与开票流程已稳定，应新建独立 `sales-manage`（Go + MySQL）承接全部本地写路径，并把列表筛选上收到后端 `qp-*`。

## What Changes

- 新建 `apps/sales-manage`：Gin + GORM + MySQL，独立库（建议 `sales_manage`，端口建议 `8083`）
- 提供销售单与开票单 HTTP API，前缀 `/api/v1/sale/...`（有意区别于 stock-manage 的 `/api/v1/sales-orders`）
- 列表 / 开票候选筛选由后端 `qp-*` + `page` / `pageSize` 处理
- `sales-front` 的 `salesOrderApi` / `invoiceDocApi` 改为 HTTP；页面与打印尽量不动
- **BREAKING（销售开票工具）**：停用 IndexedDB 读写真源（空库直切；IndexedDB→MySQL 导入脚本 **第二期**）
- 本变更 **不** 对接 stock-center / stock-manage，**不** 复用库存域销售单模型

## Capabilities

### New Capabilities

- `sale-orders`：销售单持久化与 `/api/v1/sale/orders` CRUD / 列表 `qp-*`
- `sale-invoice-docs`：开票单持久化、候选查询、创建/作废事务与 `/api/v1/sale/invoice-docs`
- `sales-front-http`：前端 Api 切 HTTP、列表编 `qp-*`、移除 Dexie 写路径

### Modified Capabilities

（无既有 openspec spec）

## Impact

- 新增：`apps/sales-manage/**`（cmd / internal / pkg、README、本地启动）
- 修改：`apps/sales-front/src/services/*Api.ts`、列表页筛选参数、测试 setup；移除或停用 `src/storage/*` 与 dexie 依赖
- 文档：`docs/domain-sales.md`、`docs/api-sales-storage.md`；kb `sales-manage-api.md` / `local-first-to-go-mysql.md` 落地后改状态
- 参考：`apps/stock-center/internal/pkg/qp`、`apps/stock-manage` 分层；约定见 `~/Mycodes/kb/domains/personal-go/sales-manage-api.md`
