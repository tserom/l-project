# 销售单接口与持久化设计

> 更新：2026-08-08（IndexedDB → `sales-manage` HTTP）  
> 配套领域文档：[`domain-sales.md`](./domain-sales.md)  
> 后端：`apps/sales-manage`；前端 Api：`apps/sales-front/src/services/*`  
> kb：`domains/personal-go/sales-manage-api.md`；过渡方法：`local-first-to-go-mysql.md`

---

## 1. 分层约定

```
页面 / 组件
    │  只调用 services（禁止直连 fetch URL 散落）
    ▼
*Api.ts（编 qp-*、hydrate、错误抛出）
    │
    ▼
httpClient → Vite proxy /api → sales-manage :8083
    │
    ▼
MySQL（库 sales_manage）头表 + 行表
```

| 层 | 做什么 | 不做什么 |
|----|--------|----------|
| Api | 拼 query、调用 HTTP、hydrate `needInvoice` | 不碰 React |
| httpClient | 解包 `{ code, message, data }` | 不含业务字段白名单 |
| 后端 service | 校验、合计、开票事务 | — |

开票创建/作废：**单次 HTTP**，事务在 Go 内完成。

---

## 2. HTTP 路由（sales-manage）

前缀：`/api/v1/sale`

| 前端动作 | 方法 | 路径 |
|----------|------|------|
| `listSalesOrders` | GET | `/orders`（`qp-*` + `page`/`pageSize`） |
| `getSalesOrder` | GET | `/orders/:id` |
| `createSalesOrder` | POST | `/orders` |
| `updateSalesOrder` | PUT | `/orders/:id` |
| `removeSalesOrder` | DELETE | `/orders/:id`（有开票挂接则 400） |
| `listInvoiceCandidates` | GET | `/orders/invoice-candidates` |
| `listInvoiceDocs` | GET | `/invoice-docs` |
| `getInvoiceDoc` | GET | `/invoice-docs/:id` |
| `createInvoiceDoc` | POST | `/invoice-docs` |
| `voidInvoiceDoc` | POST | `/invoice-docs/:id/void` |

**不是** stock-manage 的 `/api/v1/sales-orders`。

### 2.1 销售单列表 qp 白名单

| qp | 说明 |
|----|------|
| `qp-customerName-like` | 客户模糊 |
| `qp-orderNo-eq` / `in` / `like` | 单号 |
| `qp-orderDate-gte` / `lte` | 日期 |
| `qp-warehouseName-eq` | 仓库（可选） |

非法 field/operator → HTTP 400。

### 2.2 开票列表 qp

| qp | 说明 |
|----|------|
| `qp-status-eq` | `saved` / `voided` |
| `qp-invoiceNo-like` | 开票单号 |

---

## 3. 响应信封

```json
{ "code": 0, "message": "ok", "data": { } }
```

列表 `data`：`{ list, total, page, pageSize }`。  
业务错误：`code != 0`，`message` 为中文文案（如「销售单不存在」「明细已开票，请重新加载」）。

---

## 4. 表映射

| 概念 | MySQL |
|------|--------|
| 销售单 | `sales_order` + `sales_order_line` |
| 开票单 | `invoice_doc` + `invoice_doc_line` |
| 行挂接 | `sales_order_line.invoice_doc_id` |

对外 id：字符串 UUID。HTTP JSON 仍嵌套 `lines[]`，对齐 `types/*.ts`。

---

## 5. 前端文件

| 路径 | 说明 |
|------|------|
| `src/services/httpClient.ts` | fetch + 信封 |
| `src/services/salesOrderApi.ts` | 销售单 |
| `src/services/invoiceDocApi.ts` | 开票（含纯函数 `listInvoiceCandidatesFromOrders` 供单测） |
| `vite.config.ts` | `/api` → `127.0.0.1:8083` |

---

## 6. 历史：IndexedDB（已退役）

阶段一曾用 Dexie 库名 `sales-front`（表 `salesOrders` / `invoiceDocs`）。  
导入脚本第二期；空库直切即可。旧说明见 git 历史与 kb `indexeddb-dexie.md`。
