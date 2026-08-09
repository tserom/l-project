# 销售单领域设计（sales-front）

> 录入日期：2026-08-03；**2026-08-08** 更新为 MySQL + `sales-manage`  
> 定位：**当前实现真源摘要**；早期决策稿见文末「历史设计稿」  
> 代码路径：`apps/sales-front` + `apps/sales-manage`  
> 接口：[`api-sales-storage.md`](./api-sales-storage.md)  
> kb：`domains/personal-go/sales-manage-api.md`

---

## 1. 产品定位

本地工具：**销售单录单** + **出库单样式预览打印** + **开票单汇总**。

| 项 | 事实 |
|----|------|
| 运行形态 | 前端 Vite + React 19 + antd 6；后端 Gin + GORM + MySQL |
| 持久化 | `sales-manage`（:8083，库 `sales_manage`）；**不再**用 IndexedDB |
| 与库存仓 | 与 `stock-center` / `stock-manage` **无业务打通**（勿混用 `/sales-orders`） |
| 分发 | `make pack-sales-front-windows`（静态包仍需可访问的后端） |
| 开发 | `make dev-sales-manage` + `make dev-sales-front` → `http://localhost:5175` |

**核心决策（已落地）**：业务实体是**销售单**；打印只是套用纸质**出库单**版式。本应用内没有独立的「出库/发货」实体。

顶栏产品名：「销售开票」；导航：销售单 / 开票单。

---

## 2. 架构

```
Pages（列表 / 编辑 / 打印 / 开票）
        │
        ▼
  salesOrderApi / invoiceDocApi → httpClient → /api/v1/sale/*
        │
        ▼
  sales-manage（Gin）→ MySQL
        │
Print：SalesOrder + PrintProfile（常量）→ OutboundSlip DOM + @media print
```

| 层 | 职责 | 主要路径 |
|----|------|----------|
| pages | 路由页面 | `apps/sales-front/src/pages/*` |
| components | 表单、明细表、打印件 | `src/components/*` |
| services | HTTP Api | `src/services/*Api.ts` |
| types | 领域类型 | `src/types/*` |
| config | 打印档案常量 | `src/config/printProfile.ts` |
| backend | 校验、合计、开票事务 | `apps/sales-manage/internal/*` |

---

## 3. 领域模型

### 3.1 销售单 `SalesOrder`

无单据状态机（无草稿/已出库等）。

| 字段 | 说明 |
|------|------|
| `id` | 本地 UUID |
| `orderNo` | 单据号；新建默认 `YYYYMMDD` + 时间戳末四位，可改 |
| `orderDate` | 日期字符串 |
| `customerName` | 客户（必填） |
| `warehouseName` | 仓库（新建默认「中大慧科」） |
| `deliveryType` | 出库类型自由文本（新建默认「提货」） |
| `remark?` | 单据备注 |
| `outstandingBalance?` | 往来欠款（手工；空则打印不展示） |
| `lines` | 明细（嵌在单据内，无独立表） |
| `totalQuantity` / `totalAmount` | 写入时按行重算 |
| `createdAt` / `updatedAt` | ISO 时间戳 |

### 3.2 销售明细 `SalesOrderLine`

| 字段 | 说明 |
|------|------|
| `materialName` / `spec` / `unit` | 物资、规格、单位（新建行默认单位 `kg`） |
| `quantity` / `unitPrice` / `amount` | 金额默认 `round2(qty × price)` |
| `lineRemark?` | 行备注 |
| `needInvoice` | 是否需开票，默认 `false` |
| `invoiceDocId?` | 已挂到未作废开票单时有值 |

校验（API 层）：客户非空、至少一行、数量/单价 ≥ 0。

### 3.3 开票单 `InvoiceDoc`

| 字段 | 说明 |
|------|------|
| `invoiceNo` | 本地号，形如 `KP` + 销售单号算法 |
| `status` | `saved` \| `voided` |
| `filter*` | 创建时筛选快照（客户 / 日期起止 / 单据号列表） |
| `lines` | 开票明细快照（含来源销售单字段） |
| `totalQuantity` / `totalAmount` | 合计 |
| `voidedAt?` | 作废时间 |

开票明细含 `salesOrderId` / `salesOrderLineId` + 物资字段 + 来源 `orderNo` / `orderDate` / `customerName` / `warehouseName` / `deliveryType`。

**待开票候选**：`needInvoice === true` 且无 `invoiceDocId`，再叠加可选筛选（客户包含、日期含当日、单据号多值精确）。

边界：

- 不拆数量；整行进开票单
- 保存时行已被占用 → 报错
- 作废 → `status=voided`，清空相关销售明细的 `invoiceDocId`
- MVP 无开票打印、无税务对接

### 3.4 打印档案 `PrintProfile`（不进 IndexedDB）

常量在 `src/config/printProfile.ts`，当前示例公司为「淄博钰鑫不锈钢有限公司」，标题后缀「出库单」；含农行卡号、地址、电话、质量声明、纸张约 190×140 mm、表格最少 10 行（不足补空行）。

| 属于销售单 | 属于 PrintProfile |
|------------|-------------------|
| 日期、单号、客户、仓库、出库类型、明细、合计、备注、往来欠款 | 公司名、标题「出库单」、卡号、地址电话、质量声明、纸张 mm |

---

## 4. 持久化

| 项 | 事实 |
|----|------|
| DB 名 | `sales-front` |
| v1 | `salesOrders: id, orderNo, updatedAt` |
| v2 | + `invoiceDocs: id, invoiceNo, status, updatedAt` |
| localStorage | 未使用 |
| 导入/导出 | 无 |
| 种子数据 | 已移除（2026-07-29）；旧浏览器残留需手动删 |

**风险**：清除站点数据会丢单；分发包数据按机器/浏览器隔离，不跨机同步。

---

## 5. 页面与流程

### 5.1 路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | redirect | → `/orders` |
| `/orders` | 列表 | 筛选、删除、进编辑/打印 |
| `/orders/new` | 编辑 | 新建 |
| `/orders/:id` | 编辑 | 编辑即详情（无只读详情页） |
| `/orders/:id/print` | 打印预览 | `window.print()` |
| `/invoices` | 开票列表 | 查看、作废 |
| `/invoices/new` | 创建开票 | 筛候选 → 删行 → 保存 |
| `/invoices/:id` | 开票详情 | 明细 / 合并视图 |

### 5.2 销售单列表筛选

| 条件 | 规则 |
|------|------|
| 单据号 | 多值（回车 / 逗号 / 空格），精确匹配任一 |
| 日期 | 闭区间 |
| 客户 | 名称包含 |

底部对当前筛选结果求金额合计。

### 5.3 打印

- `OutboundSlip` + 默认 `PrintProfile`
- 工具条 / 顶栏 `.no-print`；浏览器原生打印
- 建议用户：关页眉页脚；边距默认或最小

### 5.4 开票创建摘要

1. 销售编辑勾选明细「需开票」  
2. 开票新建：可选筛选项 → 加载候选 → 可删行 → 保存回写 `invoiceDocId`  
3. 详情支持「明细视图 / 合并视图」（同物资+规格+单价合并数量）

---

## 6. 与库存 openspec 的边界

`openspec/changes/stainless-steel-inventory/` 中的 `sales_order` / 出库扣减属于 **stock-\*** 库存主链路，与本 `sales-front` **未打通**。

本应用当前是独立本地工具；对接库存若做，属于后续演进，需另开变更，不要默认本设计已覆盖。

---

## 7. 明确非目标（当前仍成立）

- 真实 HTTP / MySQL / 登录权限
- 销售单状态机
- 打印模板可视化配置 UI
- JSON 导入导出备份 UI
- PDF 库出票；针式齿孔/复写联纹理
- 税务发票号 / 税局接口
- 与 stock 一键包混打（库存 pack 与销售 pack 分离）

---

## 8. 演进方向（未实现）

| 方向 | 备注 |
|------|------|
| 换真 API | 替换 services / storage，保留页面与打印 |
| PrintProfile 可配置 | 现为硬编码常量 |
| 导入导出 | 缓解清站点丢数据 |
| 对接库存出库 | 需对齐 openspec 与单据语义 |

---

## 9. 历史设计稿与变更记录

| 文档 | 用途 |
|------|------|
| `docs/superpowers/specs/2026-07-24-sales-order-app-design.md` | MVP 决策与模型（批准稿） |
| `docs/superpowers/plans/2026-07-24-sales-order-app.md` | 实现计划 |
| `docs/superpowers/specs/2026-07-28-invoice-doc-design.md` | 开票单方案 |
| `apps/sales-front/docs/CHANGELOG.md` | 面向业务的变更日志 |
| `apps/sales-front/README.md` | 启动 / 打包 / 注意点 |

若本文与源码冲突，**以源码为准**，并回改本文标注日期。
