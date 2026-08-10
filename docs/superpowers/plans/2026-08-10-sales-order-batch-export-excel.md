# 销售单批量导出 Excel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 销售单列表支持「批量导出」：勾选优先，否则按已应用筛选分页拉全，前端生成汇总字段 `.xlsx`。

**Architecture:** L0（`exportSalesOrdersXlsx`）纯映射+SheetJS 下载；L3（`exportSalesOrdersToExcel`）决定数据源并循环 `listSalesOrders`；`OrderListPage` 勾选与按钮。不改后端。

**Tech Stack:** React 19、antd 6、`xlsx`、Vitest、现有 `listSalesOrders` + `qp-*`。

**Spec:** `docs/superpowers/specs/2026-08-10-sales-order-batch-export-excel-design.md`

## Global Constraints

- 导出列仅：单据号、日期、客户、仓库、出库类型、数量合计、金额合计
- 勾选 ≥1 → 导勾选；否则按 `applied` 筛选拉全（`pageSize=500`）
- 空数据：不下载，UI `message.warning`
- 文件名：`销售单_YYYYMMDD_HHmmss.xlsx`
- 禁止 BsSula；不改 `sales-manage`

## File map

| File | Responsibility |
|------|----------------|
| `apps/sales-front/package.json` | 依赖 `xlsx` |
| `apps/sales-front/src/utils/exportSalesOrdersXlsx.ts` | L0：行映射 + 写文件下载 |
| `apps/sales-front/src/utils/exportSalesOrdersXlsx.test.ts` | L0 行映射单测 |
| `apps/sales-front/src/services/salesOrderApi.ts` | L3 `exportSalesOrdersToExcel` + 入口索引 |
| `apps/sales-front/src/pages/OrderListPage.tsx` | 勾选、导出按钮、loading |
| `docs/domain-sales.md` | 导入/导出说明 |

---

### Task 1: L0 行映射 + xlsx 下载

**Files:**
- Modify: `apps/sales-front/package.json`（`pnpm add xlsx`）
- Create: `apps/sales-front/src/utils/exportSalesOrdersXlsx.ts`
- Create: `apps/sales-front/src/utils/exportSalesOrdersXlsx.test.ts`

**Interfaces:**
- Produces:
  - `salesOrdersToSheetRows(orders: SalesOrder[]): (string | number)[][]` — 含表头
  - `buildSalesOrderExportFilename(now?: Date): string`
  - `downloadSalesOrdersXlsx(orders: SalesOrder[], now?: Date): void`

- [x] **Step 1: 安装依赖**
- [x] **Step 2: 写失败测试**
- [x] **Step 3: 跑测确认失败**
- [x] **Step 4: 实现 L0**
- [x] **Step 5: 跑测通过**

---

### Task 2: L3 `exportSalesOrdersToExcel`

**Files:**
- Modify: `apps/sales-front/src/services/salesOrderApi.ts`

**Interfaces:**
- Consumes: `listSalesOrders`, `downloadSalesOrdersXlsx`
- Produces:

```ts
export type ExportSalesOrdersResult =
  | { ok: true }
  | { ok: false; reason: 'empty' }

export async function exportSalesOrdersToExcel(input: {
  selectedOrders?: SalesOrder[]
  query?: ListSalesOrdersQuery
}): Promise<ExportSalesOrdersResult>
```

- [x] **Step 1: 更新入口索引并实现**
- [x] **Step 2: 类型检查**

---

### Task 3: 列表页 UI

**Files:**
- Modify: `apps/sales-front/src/pages/OrderListPage.tsx`

**Interfaces:**
- Consumes: `exportSalesOrdersToExcel`

- [x] **Step 1: 状态与勾选**
- [x] **Step 2: 导出 handler**
- [x] **Step 3: 按钮**

---

### Task 4: 文档

**Files:**
- Modify: `docs/domain-sales.md`

- [x] **Step 1:** 将 `| 导入/导出 | 无 |` 改为 `| 导入/导出 | 列表汇总 xlsx（前端生成；勾选优先） |`

---

### Task 5: 验证

- [x] **Step 1:** `cd apps/sales-front && pnpm test`
- [x] **Step 2:** `cd apps/sales-front && pnpm exec tsc -b --pretty false`
- [ ] **Step 3:** 手动：列表勾选导出 / 无勾选按筛选导出 / 空列表警告
