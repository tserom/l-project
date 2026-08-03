# 销售单接口与 IndexedDB 设计

> 录入日期：2026-08-03（按 `apps/sales-front` 现码）  
> 配套领域文档：[`domain-sales.md`](./domain-sales.md)  
> 代码：`apps/sales-front/src/services/*`、`src/storage/*`  
> kb 通用：IndexedDB 用法 `@~/Mycodes/kb/domains/personal-react/indexeddb-dexie.md`；  
> 迁 Go+MySQL 方法 `@~/Mycodes/kb/domains/workflow/local-first-to-go-mysql.md`

---

## 1. 分层约定

```
页面 / 组件
    │  只调用 services（禁止直接 import db / Repository）
    ▼
*Api.ts（校验、合计、事务编排、错误文案）
    │
    ▼
*Repository.ts（单表 CRUD）  或  跨表时 Api 内直接 db.transaction
    │
    ▼
Dexie → 浏览器 IndexedDB（库名 sales-front）
```

| 层 | 做什么 | 不做什么 |
|----|--------|----------|
| Api | 业务规则、重算合计、hydrate、跨表事务 | 不碰 React / antd |
| Repository | `get` / `put` / `delete` / 按索引排序列表 | 不做校验、不算金额 |
| db.ts | schema、version、表声明 | 不含业务逻辑 |

例外：`createInvoiceDoc` / `voidInvoiceDoc` 需同时改 `salesOrders` 与 `invoiceDocs`，在 Api 内用 `db.transaction`，不经 Repository 拼事务。

换真 HTTP 时：保持 Api 函数签名，把 Repository 换成 `fetch`；跨表事务改由后端保证。  
完整过渡清单见 kb：`domains/workflow/local-first-to-go-mysql.md`。

---

## 2. IndexedDB（Dexie）怎么用

### 2.1 库与版本

定义：`src/storage/db.ts`

| 项 | 值 |
|----|-----|
| 库名 | `sales-front` |
| 封装 | Dexie 类 `SalesFrontDB` + 单例 `db` |
| v1 | 表 `salesOrders`，主键 `id`，索引 `orderNo`, `updatedAt` |
| v2 | 增表 `invoiceDocs`，主键 `id`，索引 `invoiceNo`, `status`, `updatedAt` |

```ts
// 结构示意（与现码一致）
this.version(1).stores({ salesOrders: 'id, orderNo, updatedAt' })
this.version(2).stores({
  salesOrders: 'id, orderNo, updatedAt',
  invoiceDocs: 'id, invoiceNo, status, updatedAt',
})
```

Dexie 首次打开会建库；已有 v1 数据的浏览器升级到 v2 时**只加表**，不迁销售单行内字段。`needInvoice` / `invoiceDocId` 嵌在 `lines` JSON 里，无独立列，故无需 schema 迁移脚本。

### 2.2 存什么、怎么查

| 对象仓库 | 一条记录是什么 | 明细怎么存 |
|----------|----------------|------------|
| `salesOrders` | 整张 `SalesOrder` | `lines[]` **嵌在单据文档内** |
| `invoiceDocs` | 整张 `InvoiceDoc` | `lines[]` 同样嵌入 |

索引只服务顶层字段（单号、更新时间、开票状态）。  
**列表筛选、待开票候选**：先 `listAll()` 拉全量，再在内存过滤（数据量按单机本地工具假设）。IndexedDB 侧不做复合查询。

常用 Repository 写法：

```ts
// 按 updatedAt 新→旧
db.salesOrders.orderBy('updatedAt').reverse().toArray()

// 主键读写
db.salesOrders.get(id)
db.salesOrders.put(order)   // 有则覆盖，无则插入
db.salesOrders.delete(id)
```

### 2.3 事务

仅开票写路径使用显式读写事务：

```ts
await db.transaction('rw', db.salesOrders, db.invoiceDocs, async () => {
  // 回写销售明细 invoiceDocId + put 开票单
})
```

销售单单表 CRUD 用单次 `put` / `delete`，无跨表事务。

### 2.4 生命周期与风险

| 点 | 说明 |
|----|------|
| 何时创建 | 首次 `import { db }` 并发生读写时由 Dexie 打开 |
| 数据位置 | 同源（协议+域名+端口）下的浏览器配置文件 |
| 清站点数据 | 整库丢失；MVP 无导出 |
| 多标签页 | Dexie 默认同库；无应用层锁；开票靠保存时再校验 `invoiceDocId` |
| DevTools | Application → IndexedDB → `sales-front` 可查看两表 |

Repository 里另有 `count` / `putMany`（销售单），**现码无调用方**（原种子数据用过后遗留，待核实是否删除）。

---

## 3. 销售单 Api

文件：`src/services/salesOrderApi.ts`  
存储：仅 `salesOrderRepository` → `salesOrders`

| 函数 | 入参 | 返回 | 行为摘要 |
|------|------|------|----------|
| `listSalesOrders()` | — | `SalesOrder[]` | 全量，按 `updatedAt` 降序；hydrate `needInvoice` |
| `getSalesOrder(id)` | id | `SalesOrder` | 无则抛「销售单不存在」 |
| `createSalesOrder(input)` | `SalesOrderInput` | `SalesOrder` | 校验 → `crypto.randomUUID()` → 重算行金额与合计 → `put` |
| `updateSalesOrder(id, input)` | id + input | `SalesOrder` | 校验 → 保留 `createdAt` → 用旧行补全缺失的 `invoiceDocId` → `put` |
| `removeSalesOrder(id)` | id | `void` | `delete`（不级联开票单；待核实业务是否应限制已开票删除） |

### 3.1 `SalesOrderInput`

即 `SalesOrder` 去掉由 Api 写入的：`id`（新建可省略）、`totalQuantity`、`totalAmount`、`createdAt`、`updatedAt`。调用方提供表头 + `lines`。

### 3.2 写入时规范化

- 行 `amount`：一律 `calcLineAmount(quantity, unitPrice)`（不信任表单自行金额）
- `needInvoice`：`Boolean(...)`
- 更新时 `invoiceDocId`：入参未带则沿用原行，避免编辑冲掉开票挂接
- 表头 `customerName`：`trim`
- 校验失败：`throw new Error('…')`（客户空、无明细、数量/单价为负）

页面侧：列表筛选用 `utils/filterSalesOrders`，**不是** Api 参数。

---

## 4. 开票单 Api

文件：`src/services/invoiceDocApi.ts`

| 函数 | 存储涉及 | 行为摘要 |
|------|----------|----------|
| `listInvoiceCandidates(filter)` | 读 `salesOrders` | 内存筛「需开票且未挂开票单」+ 可选客户/日期/单号 |
| `listInvoiceDocs()` | `invoiceDocs` | 全量，`updatedAt` 降序 |
| `getInvoiceDoc(id)` | 同上 | 无则抛「开票单不存在」 |
| `createInvoiceDoc(input)` | **事务**写两表 | 见下 |
| `voidInvoiceDoc(id)` | **事务**写两表 | 见下 |

纯函数（可单测）：`listInvoiceCandidatesFromOrders(orders, filter)` —— 不访问 DB。

### 4.1 `CreateInvoiceDocInput`

```ts
{
  filterCustomerName?: string
  filterDateFrom?: string
  filterDateTo?: string
  filterOrderNos?: string[]
  lines: InvoiceDocLine[]   // 至少一行；可先删候选再提交
}
```

### 4.2 创建事务（顺序）

1. 校验 `lines.length > 0`
2. 生成 `docId`、规范化行 id
3. 事务内对每一行：
   - 取销售单；找 `salesOrderLineId`
   - 必须 `needInvoice` 且当前无 `invoiceDocId`，否则抛错（提示重新加载）
   - `put` 销售单，该行写入 `invoiceDocId: docId`
4. `put` 开票单：`invoiceNo = 'KP' + generateOrderNo()`，`status: 'saved'`，合计重算，筛选条件留档
5. 事务外 `getInvoiceDoc` 返回

### 4.3 作废事务

1. 已是 `voided` → 抛「开票单已作废」
2. 事务内：对开票明细对应销售行，仅当 `invoiceDocId === 本单 id` 时去掉该字段
3. 开票单 `status: 'voided'`，写 `voidedAt` / `updatedAt`
4. 销售单缺失则跳过该行（不阻断作废）

---

## 5. 调用关系速查

```
OrderListPage / OrderEditPage / OrderPrintPage
  → salesOrderApi.*
      → salesOrderRepository.* → db.salesOrders

InvoiceListPage / InvoiceCreatePage / InvoiceDetailPage
  → invoiceDocApi.*
      → list*：orderRepo / invoiceRepo
      → create / void：db.transaction(salesOrders, invoiceDocs)
```

打印页只读 `getSalesOrder` + 常量 `defaultPrintProfile`，不写库。

---

## 6. 错误约定

| 场景 | 典型 `Error.message` |
|------|----------------------|
| 销售单缺失 | `销售单不存在` |
| 客户/明细非法 | `客户不能为空` / `至少需要一行明细` / 数量或单价负数 |
| 开票单缺失 | `开票单不存在` |
| 开票行冲突 | `明细已开票，请重新加载：…` / 未勾选需开票 / 销售单或明细不存在 |
| 重复作废 | `开票单已作废` |

页面用 `try/catch` + antd `message` 展示即可；无统一错误码枚举。

---

## 7. 测试与演进

- 单测：`*.test.ts` 覆盖 Api / 过滤 / 合并视图；测试环境用 `fake-indexeddb`（见 `src/test/setup.ts`）
- 演进换后端：保持上表函数名与返回类型；IndexedDB 仅作离线兜底时再另开设计
- 若以后要按 `orderNo` 在 DB 侧查重：已有索引，可在 Repository 加 `where('orderNo').equals(...)`，列表筛选仍可先内存

---

## 8. 相关文件

| 路径 | 说明 |
|------|------|
| `src/storage/db.ts` | Dexie schema |
| `src/storage/salesOrderRepository.ts` | 销售单表访问 |
| `src/storage/invoiceDocRepository.ts` | 开票单表访问 |
| `src/services/salesOrderApi.ts` | 销售单接口 |
| `src/services/invoiceDocApi.ts` | 开票单接口 |
| `src/types/salesOrder.ts` / `invoiceDoc.ts` | 类型真源 |
| `docs/domain-sales.md` | 领域与页面流程 |
