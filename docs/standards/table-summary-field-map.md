# 表格 / 导出「合计」：按字段映射

**状态**：生效  
**起源**：销售单列表合计在增加勾选列后错位到「数量合计」列；导出合计行曾用空串占位。  
**适用范围**：列表 `summary`、Excel/CSV 导出末行、后端聚合汇总字段（若接口返回合计）

---

## 一句话

合计内容用 **公开字段名 / 列 key** 对齐到列；**禁止**用「第 N 列 / `colSpan={k}` / 塞空字符串占位」表达对应关系。

---

## 为什么

| 做法 | 后果 |
|------|------|
| `colSpan={6}` + 下一格写金额 | 加勾选列、插列后合计整体错位（金额跑到数量下） |
| `['合计','','','',…,金额]` | 与上同：靠下标，列一变就错 |
| `byKey.totalAmount = 金额` | 列增删、勾选占位只影响遍历顺序，**值仍挂在字段上** |

前后端同一原则：协议与 UI 都以 **field** 为契约，不以「数组下标」为契约。

---

## MUST

### 前端（列表 summary）

1. 列定义带稳定 `key`（优先与 `dataIndex` / API 字段同名，如 `totalAmount`）。
2. 合计用 `Record<fieldKey, ReactNode>`（或同等映射）填值。
3. 渲染时 **按 `columns` 顺序** 取 `map[key]`；`rowSelection` 等无业务 key 的列单独空单元格占位，**不要**把业务合计写进占位列的魔法下标。

示意：

```tsx
const summaryByColumnKey: Partial<Record<string, ReactNode>> = {
  orderNo: `合计（本页 ${pageSize} / 共 ${total}）`,
  totalAmount: <strong>{formatAmount(pageAmount)}</strong>,
}

// summary 行：勾选列空单元格 + columns.map → summaryByColumnKey[key]
```

### 前端（导出 Excel / CSV）

1. 导出列定义为 `{ key, title }[]`，与列表公开字段一致。
2. 数据行、合计行都经「key → 单元格」生成；合计例如：

```ts
cellsFromByKey({
  orderNo: '合计',
  totalAmount: formatAmount(sum),
})
```

### 后端（若返回或生成合计）

1. JSON / 导出列：合计值挂在 **公开字段名** 上（如 `totalAmount`），不要 `totals[6]` 或「第 7 列是金额」。
2. 服务端生成 Excel 时同样按字段写列，禁止仅按列序号硬编码业务含义（列序可变，字段名是契约）。
3. 列表筛选等仍遵守 `qp-<field>-<operator>`；合计字段名与列表 VO 一致。

---

## MUST NOT

- 用纯数字下标表达「金额在第几列」作为唯一真源（注释里可写示例下标，实现必须以 field 为准）。
- 为对齐而长期依赖易碎的 `colSpan` 常数（除非产品明确要求「文案横跨多列」，且横跨范围也由 **起止 field key** 推导，而不是手写数字）。
- 导出合计行用一串 `''` 占位凑列数。

---

## 可选：横跨文案

若需要「合计」文案视觉上横跨多列：用 **起止字段 key** 计算合并（例如从 `orderNo` 合并到 `deliveryType` 之前），不要写死 `colSpan={7}`。无强需求时，文案落在 `orderNo`（或首列业务 key）即可。

---

## 自检 / CR

- [ ] 合计金额是否写在 `totalAmount`（或约定字段）映射里，而不是「倒数第二格」？
- [ ] 增加勾选列 / 中间插列后，金额是否仍在金额列下？
- [ ] 导出与列表的字段 key 是否一致？
- [ ] 后端若有汇总，响应/导出是否用字段名而非下标？

---

## 变更记录

| 日期 | 说明 |
|------|------|
| 2026-08-10 | 初版：从个人仓销售单合计/导出错位晋升 |
