# 开票单设计（销售明细需开票 + 开票汇总）

**日期**：2026-07-28  
**状态**：已确认（方案 1），进入实现  
**范围**：`apps/sales-front` 本地 IndexedDB；无后端

---

## 1. 目标

- 销售单**明细行**标记是否「需开票」（默认否）
- 新建**开票单**：按可选客户、日期范围、单据号筛出待开票明细，先全部带入，可删行后保存
- 保存后明细挂到开票单（已开票）；**作废**开票单后明细重新可开
- 开票单行末展示来源销售单信息（单号、日期、客户等）

## 2. 已确认决策

| 点 | 选择 |
|----|------|
| 架构 | 开票单实体 + 明细 `needInvoice` + `invoiceDocId` |
| 客户 | 可选；可不选，一张单可混多客户 |
| 入单 | 条件命中全部带入，可删行再保存 |
| 需开票默认 | 不勾选 |
| 作废 | 清空相关行的 `invoiceDocId`，状态作废 |

## 3. 数据模型

### 3.1 销售明细扩展

```ts
SalesOrderLine {
  // …原有字段
  needInvoice: boolean       // 默认 false
  invoiceDocId?: string      // 有值且对应开票单未作废 → 已开票
}
```

### 3.2 开票单

```ts
InvoiceDoc {
  id: string
  invoiceNo: string          // 本地生成
  status: 'saved' | 'voided'
  // 创建时所用筛选（留档）
  filterCustomerName?: string
  filterDateFrom?: string
  filterDateTo?: string
  filterOrderNos?: string[]
  lines: InvoiceDocLine[]
  totalQuantity: number
  totalAmount: number
  createdAt: string
  updatedAt: string
  voidedAt?: string
}

InvoiceDocLine {
  id: string
  salesOrderId: string
  salesOrderLineId: string
  // 明细业务字段（快照）
  materialName, spec, unit, quantity, unitPrice, amount, lineRemark?
  // 行末展示的单据信息（快照）
  orderNo, orderDate, customerName, warehouseName, deliveryType
}
```

### 3.3 待开票候选

```
line.needInvoice === true
且 !line.invoiceDocId
再 AND 可选：客户包含 / 日期范围 / 单据号多值精确
```

## 4. 页面与流程

```
/orders …（编辑页明细增加「需开票」勾选）

/invoices                 开票单列表（作废、查看）
/invoices/new             创建：填筛选 → 加载候选 → 删行 → 保存
/invoices/:id             查看（已作废只读）
```

导航：顶栏增加「开票单」入口。

### 创建流程

1. 输入可选：客户、日期范围、单据号（多值）
2. 「加载明细」→ 拉齐待开票行，带入表格
3. 可删行；底部金额合计
4. 保存 → 写 `InvoiceDoc`，回写各销售明细 `invoiceDocId`
5. 列表可「作废」→ 清 `invoiceDocId`，`status=voided`

## 5. 边界

- 保存时若某行已被其它开票单占用 → 报错并提示刷新
- 不作废不可重复开同一明细
- 不拆数量开票（整行进开票单）
- MVP 不做开票单打印/税务字段

## 6. 非目标

- 真实税务接口、发票号对接税局
- 多用户并发锁（单机浏览器）
- 后端 API

---

**关联实现**：本变更直接在 `feature/sales-front` 落地；修改记录写入 `apps/sales-front/docs/CHANGELOG.md`。
