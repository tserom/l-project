# 销售单往来欠款 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 销售单增加手工「往来欠款」字段，出库单打印页脚在电话后有值才显示，备注保留。

**Architecture:** 在 `SalesOrder.outstandingBalance?: string` 上透传；表单编辑；`OutboundSlip` 条件渲染。不做历史迁移、不做列表列、不做自动汇总。

**Tech Stack:** React + Ant Design Form、Dexie IndexedDB、Vitest、现有 `salesOrderApi` / `OutboundSlip`。

**Spec:** `docs/superpowers/specs/2026-08-03-sales-order-outstanding-balance-design.md`

## Global Constraints

- 字段名：`outstandingBalance`；中文标签：`往来欠款`
- 自由文本；trim 后空则打印不展示该段
- 历史数据完全不动（无迁移 / 回填 / 解析旧备注）
- 列表不加列

## File map

| File | Responsibility |
|------|----------------|
| `apps/sales-front/src/types/salesOrder.ts` | 类型字段 |
| `apps/sales-front/src/services/salesOrderApi.ts` | `buildOrder` 透传 |
| `apps/sales-front/src/services/salesOrderApi.test.ts` | 持久化往返 |
| `apps/sales-front/src/components/OrderForm.tsx` | 表单项 + `toFormValues` |
| `apps/sales-front/src/pages/OrderEditPage.tsx` | 新建默认 / 加载 / 保存 |
| `apps/sales-front/src/components/print/OutboundSlip.tsx` | 页脚条件渲染 |

---

### Task 1: 类型 + API 透传 + 测试

**Files:**
- Modify: `apps/sales-front/src/types/salesOrder.ts`
- Modify: `apps/sales-front/src/services/salesOrderApi.ts` (`buildOrder`)
- Modify: `apps/sales-front/src/services/salesOrderApi.test.ts`

**Interfaces:**
- Produces: `SalesOrder.outstandingBalance?: string`；`SalesOrderInput` 经 Omit 自动带上

- [x] **Step 1: 在 `SalesOrder` 增加字段**

```ts
remark?: string
/** 往来欠款（手工填写；空表示未填，打印时不展示） */
outstandingBalance?: string
lines: SalesOrderLine[]
```

- [x] **Step 2: `buildOrder` 透传**

```ts
remark: input.remark,
outstandingBalance: input.outstandingBalance,
lines,
```

- [x] **Step 3: 测试 — 创建后读回字段**

在 `salesOrderApi.test.ts` 增加：

```ts
it('persists outstandingBalance', async () => {
  const created = await createSalesOrder({
    ...validInput,
    outstandingBalance: '100',
  })
  expect(created.outstandingBalance).toBe('100')
  expect((await getSalesOrder(created.id)).outstandingBalance).toBe('100')
})
```

- [x] **Step 4: 跑测**

Run: `cd apps/sales-front && pnpm test -- src/services/salesOrderApi.test.ts`

Expected: PASS

---

### Task 2: 表单 + 编辑页

**Files:**
- Modify: `apps/sales-front/src/components/OrderForm.tsx`
- Modify: `apps/sales-front/src/pages/OrderEditPage.tsx`

**Interfaces:**
- Consumes: `outstandingBalance?: string` on order / input
- Produces: form values include `outstandingBalance`

- [x] **Step 1: `OrderFormValues` + 表单项（备注旁）**

```ts
remark?: string
outstandingBalance?: string
lines: SalesOrderLine[]
```

在备注 `Col` 后增加：

```tsx
<Col xs={24} md={8}>
  <Form.Item name="outstandingBalance" label="往来欠款">
    <Input />
  </Form.Item>
</Col>
```

`toFormValues` 的 input 类型增加 `outstandingBalance?: string`（spread 已覆盖）。

- [x] **Step 2: `OrderEditPage` 新建默认、加载、保存映射**

- 新建：`outstandingBalance: ''`
- `toFormValues({ ..., outstandingBalance: order.outstandingBalance, ... })`
- save input：`outstandingBalance: values.outstandingBalance`

---

### Task 3: 打印页脚

**Files:**
- Modify: `apps/sales-front/src/components/print/OutboundSlip.tsx`

- [x] **Step 1: 电话后条件渲染**

```tsx
<div className="outbound-slip__footer">
  <span>地址：{profile.address}</span>
  <span>电话：{profile.phone}</span>
  {order.outstandingBalance?.trim() ? (
    <span>往来欠款：{order.outstandingBalance.trim()}</span>
  ) : null}
  <span>备注：{order.remark ?? ''}</span>
</div>
```

- [x] **Step 2: 全量测试**

Run: `cd apps/sales-front && pnpm test`

Expected: PASS

- [x] **Step 3: 更新设计状态为已实现（计划完成时）**

将 spec 状态改为「已实现」并链到本 plan（可选，与实现同批）。
