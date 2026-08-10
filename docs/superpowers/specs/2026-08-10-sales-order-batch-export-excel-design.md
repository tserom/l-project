# 销售单列表「批量导出 Excel」设计

**日期**：2026-08-10  
**状态**：已实现；实现计划见 `docs/superpowers/plans/2026-08-10-sales-order-batch-export-excel.md`  
**范围**：`apps/sales-front` — 销售单列表汇总字段前端生成 `.xlsx`  
**关联**：延续 `docs/superpowers/specs/2026-07-24-sales-order-app-design.md`；列表 API 见 `apps/sales-front/src/services/salesOrderApi.ts`

---

## 1. 背景与目标

### 1.1 问题

销售单列表已支持筛选与查看，但没有导出能力；需要把当前结果落到 Excel，便于线下核对。用户后续可能做「后端生成 + OSS 时效下载」，但目前无 OSS，本轮用**前端生成 xlsx** 练手并交付可用能力。

### 1.2 目标

| 目标 | 说明 |
|------|------|
| 批量导出 | 列表页一键下载 `.xlsx` |
| 内容 | 仅列表汇总字段（一行一单） |
| 范围 | 有勾选导勾选；无勾选按当前筛选拉全再导 |
| 小样 | 前端 SheetJS，不依赖后端 export / OSS |

### 1.3 已确认决策

| 决策点 | 选择 |
|--------|------|
| 导出内容 | **A**：仅列表字段（单据号、日期、客户、仓库、出库类型、数量合计、金额合计） |
| 导出范围 | **C**：勾选优先；无勾选则按当前筛选结果 |
| 无勾选时数据量 | **B**：导出时重新请求，分页拉全匹配结果（后端单页上限 500） |
| 实现路径 | **方案 1**：纯前端生成 xlsx（接受；远期倾向方案 2 + OSS，本轮不实现） |

### 1.4 非目标

- 后端 `/export` 接口、OSS 上传、时效下载 URL
- 明细行 Sheet、多 Sheet
- 勾选跨页记忆 / 全选所有匹配行（勾选仅针对当前列表已加载行）
- 导入、模板配置、开票单导出
- 修改 `sales-manage` 的 `pageSize` 上限（沿用现有 500）

---

## 2. 交互

### 2.1 入口

- 位置：销售单列表，「新建销售单」旁增加「批量导出」按钮（非主色即可，避免抢「新建」）。
- 点击后按钮进入 loading，防止重复点击。

### 2.2 数据范围

| 条件 | 行为 |
|------|------|
| 已勾选 ≥ 1 行 | 仅导出勾选对应的订单（取当前表格已加载数据中的选中行） |
| 未勾选 | 用**当前已应用的筛选条件**（与「查询」后的 `applied` 一致，而非未点查询的表单草稿）分页拉取全部匹配单，再导出 |
| 结果为空 | `message.warning` 提示无可导出数据，**不**下载空文件 |

### 2.3 文件

| 项 | 约定 |
|----|------|
| 格式 | `.xlsx` |
| 文件名 | `销售单_YYYYMMDD_HHmmss.xlsx` |
| Sheet | 单 Sheet，表头中文列名，一行一单 |

### 2.4 表格勾选

- `Table` 增加 `rowSelection`（checkbox）。
- 勾选状态仅用于导出；不改变删除等其它操作。
- 查询 / 重置成功刷新列表后清空勾选。

---

## 3. 导出列

| Excel 列名 | 字段 | 格式 |
|------------|------|------|
| 单据号 | `orderNo` | 原文 |
| 日期 | `orderDate` | 与列表一致（`formatOrderDate`） |
| 客户 | `customerName` | 原文 |
| 仓库 | `warehouseName` | 原文 |
| 出库类型 | `deliveryType` | 原文 |
| 数量合计 | `totalQuantity` | 与列表展示一致（`formatQuantity`） |
| 金额合计 | `totalAmount` | 与列表展示一致（`formatAmount`） |

不含「操作」列；不含明细行。末行「合计」：单据号列写「合计」，金额合计列为导出行金额之和（格式同列表）。

---

## 4. 技术分层

遵循 API L0/L1/L3 约定；页面不手写多接口串联与 xlsx 细节。

| 层 | 位置 | 职责 |
|----|------|------|
| UI | `OrderListPage.tsx` | `rowSelection`、导出按钮与 loading、组装入参（选中订单或筛选 query） |
| L3 | `exportSalesOrdersToExcel` in `salesOrderApi.ts` | 决定数据来源；无勾选时循环 `listSalesOrders`；调用 L0 写文件并下载 |
| L0 | `utils/exportSalesOrdersXlsx.ts` | 行映射、`xlsx` 生成 workbook、触发浏览器下载；**不发 HTTP** |
| 依赖 | `xlsx`（SheetJS 社区版） | 仅前端 |

### 4.1 L3 伪流程

```
exportSalesOrdersToExcel({ selectedOrders?, query? })
  1. 若 selectedOrders 非空 → rows = selectedOrders
     否则 → 按 query 分页 list（pageSize=500）直到收齐 total
  2. 若 rows 为空 → 返回 { ok: false, reason: 'empty' }（不下载）；UI 据此 message.warning
  3. 否则 L0 build + download，返回 { ok: true }
```
### 4.2 分页拉全

- 复用现有 `listSalesOrders` + `qp-*` 映射，不新增 list URL。
- `pageSize` 取 500（仓库上限）；`page` 从 1 递增，直到 `list.length` 累计 ≥ `total` 或本页为空。
- 入口索引表增加一行：`批量导出 | exportSalesOrdersToExcel | GET /api/v1/sale/orders（可能多次）`。

### 4.3 错误处理

- 列表请求失败：沿用现有错误信息，`message.error`。
- 生成/下载异常：`message.error`，结束 loading。

---

## 5. 文档与远期

- 可选：`docs/domain-sales.md` 将「导入/导出 \| 无」改为「列表汇总 xlsx（前端生成）」。
- **远期方案 2**（本轮不实现）：后端生成文件 → 上传 OSS → 返回时效下载 URL；前端可保留同一按钮与「勾选优先 / 否则按筛选」语义，仅替换 L3 取数/下载实现。

---

## 6. 测试要点

| 场景 | 期望 |
|------|------|
| 无数据点导出 | 警告，无文件 |
| 无勾选 + 有筛选 | 导出文件行数 = 筛选匹配总数（可多于当前页加载的 200） |
| 勾选部分行 | 仅含勾选行，与筛选全量无关 |
| 列与格式 | 与列表七列字段一致 |
| 文件名 | 符合 `销售单_YYYYMMDD_HHmmss.xlsx` |

单元测试优先覆盖 L0 行映射（给定 `SalesOrder[]` → AOA / sheet 数据），不必测真实文件对话框。
