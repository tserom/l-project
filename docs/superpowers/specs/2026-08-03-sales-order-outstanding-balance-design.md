# 销售单「往来欠款」设计

**日期**：2026-08-03  
**状态**：待用户确认后进入实现计划  
**范围**：`apps/sales-front` — 销售单表头手工字段 + 出库单打印页脚展示  
**关联**：延续 `docs/superpowers/specs/2026-07-24-sales-order-app-design.md`

---

## 1. 背景与目标

### 1.1 问题

纸质出库单页脚在「电话」后常需标明「往来欠款」。当前销售单只有表头 `remark`（备注），业务上常把「往来欠款：100」写进备注，导致：

- 真正备注与欠款文案混在一起
- 打印页脚只能整段输出 `备注：…`，无法固定「往来欠款」标签与展示规则

### 1.2 目标

| 目标 | 说明 |
|------|------|
| 独立字段 | 销售单增加「往来欠款」，手工核实、可改 |
| 与备注并存 | 表单与打印均保留原「备注」 |
| 打印位置 | 出库单页脚：地址 → 电话 → 往来欠款（有值才显示）→ 备注 |
| 最小改动 | 不做客户档案、不做自动汇总、不改列表列 |

### 1.3 已确认决策

| 决策点 | 选择 |
|--------|------|
| 方案 | 销售单独立可选字段（方案 1） |
| 打印页脚 | **B**：保留「备注」，另加「往来欠款」 |
| 空值 | **B**：未填（trim 后空）则不打印「往来欠款」整段 |
| 字段类型 | **B**：自由文本（如 `100`、`已结清`、`约 200`） |
| 列表 | 本轮不加「往来欠款」列 |

### 1.4 非目标

- 按客户自动汇总欠款 / 应收台账
- 开单时从客户档案带出欠款
- 数字格式强制校验、金额运算、与本单合计联动
- 列表筛选、导出、统计报表
- 迁移/解析历史备注里「往来欠款：…」文案

---

## 2. 数据模型

在 `SalesOrder` 上增加：

```ts
/** 往来欠款（手工填写；空表示未填，打印时不展示） */
outstandingBalance?: string
```

| 项 | 约定 |
|----|------|
| 存储 | IndexedDB 整对象存 `SalesOrder`；Dexie 索引不变（无需新 version 仅为加字段） |
| 空值语义 | `undefined` / 缺省 / 仅空白 → 视为未填 |
| 旧数据 | 无该字段的历史单等同未填，打印行为与现网一致 |
| API 输入 | `create` / `update` 透传 `outstandingBalance`（与 `remark` 同级） |

字段英文名用 `outstandingBalance`；UI / 打印中文标签固定为「往来欠款」。

---

## 3. UI 与打印

### 3.1 编辑表单（`OrderForm`）

- 新增 `Form.Item`：`name="outstandingBalance"`，`label="往来欠款"`，控件为普通 `Input`
- 放置：与表头「备注」同区（同排或紧挨备注），不另开区块
- 非必填；不限制字符格式
- `OrderEditPage` 保存时把 `values.outstandingBalance` 写入 API input；回填编辑时带上该字段

### 3.2 列表

本轮 **不** 增加列或筛选。

### 3.3 出库单打印（`OutboundSlip`）

页脚现结构：

```text
地址：{profile.address}    电话：{profile.phone}    备注：{order.remark}
```

改为：

```text
地址：{profile.address}    电话：{profile.phone}    [往来欠款：{value}]    备注：{order.remark ?? ''}
```

其中 `[往来欠款：{value}]` **仅当** `order.outstandingBalance?.trim()` 非空时渲染。

备注段逻辑不变（可继续打出空「备注：」）；地址 / 电话仍来自 `PrintProfile`，不改为订单字段。

CSS：沿用 `.outbound-slip__footer` 的横向 `span` 布局；有值时多一个 `span`，无需新纸张尺寸约定。

### 3.4 数据流

```
OrderForm (outstandingBalance)
    → salesOrderApi create/update
    → Repository / IndexedDB
    → getSalesOrder
    → OutboundSlip（有值才渲染「往来欠款：」）
```

无单独 DTO；全量 `SalesOrder` + `defaultPrintProfile` 与现打印链路一致。

---

## 4. 实现触点（清单）

| 区域 | 文件（预期） | 改动 |
|------|----------------|------|
| 类型 | `src/types/salesOrder.ts` | 增加 `outstandingBalance?` |
| API 组装 | `src/services/salesOrderApi.ts`（及 repository 若有显式字段拷贝） | create/update/build 透传 |
| 表单 | `src/components/OrderForm.tsx` | 表单项 |
| 编辑页 | `src/pages/OrderEditPage.tsx` | 保存 / 初始值映射 |
| 打印 | `src/components/print/OutboundSlip.tsx` | 页脚条件渲染 |
| 测试 | 现有 order / print 相关单测（若有） | 空值不打印、有值打印 |

无需改 `printProfile`、Makefile、打包脚本。

---

## 5. 验收标准

1. 新建/编辑销售单可填写、修改、清空「往来欠款」，保存后刷新仍在。
2. 打印预览：有非空白值时，电话后出现 `往来欠款：{原样文本}`；备注仍在其后。
3. 未填或仅空格：页脚不出现「往来欠款」字样；备注行为与改前一致。
4. 历史无该字段的订单可正常打开、打印，不报错。
5. 不自动计算、不依赖客户名。

---

## 6. 后续可选（不做进本变更）

- 列表展示 / 按欠款筛选
- 客户级往来台账与开单带出
- 从历史 `remark` 试解析迁到本字段
