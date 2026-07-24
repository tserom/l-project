# 销售单 App 设计（录单 + 出库单样式打印）

**日期**：2026-07-24  
**状态**：已批准；实现计划见 `docs/superpowers/plans/2026-07-24-sales-order-app.md`  
**范围**：第一版 MVP — 独立前端 `apps/sales-front`，本地持久化，无后端、无数据库服务

---

## 1. 背景与目标

### 1.1 业务场景

需要记录与展示**销售单**，并按现有纸质**出库单**样式预览、打印。样张标题为「泰州市金阳金属制品有限公司出库单」，表头含客户、仓库、出库类型、农行卡号等；其中部分字段属于公司/打印模板，不属于销售单业务本体，后续可定制。

### 1.2 产品目标

| 目标 | 说明 |
|------|------|
| 销售单 CRUD | 新建、编辑、列表查看（删除可选） |
| 预览打印 | 浏览器打印，版式尽量还原连续纸半页出库单 |
| 数据模型清晰 | 销售单字段与打印档案字段分离 |
| 可持久化 | 同机同浏览器刷新后数据仍在 |
| 可演进 | API 外形可换真后端；打印档案可改为可配置 |

### 1.3 已确认决策

| 决策点 | 选择 |
|--------|------|
| 业务 vs 打印 | **A**：业务实体是销售单；打印套用出库单样式 |
| 放置位置 | **A**：本仓新建独立前端 `apps/sales-front`，与 `stock-front` 平级，先不接库存服务 |
| 使用方式 | **A**：主要同一电脑、同一浏览器 |
| MVP 功能 | **A**：新建/编辑、列表、预览并打印（不做模板配置 UI、不做导入导出） |
| 打印保真 | **B**：尽量还原连续纸半页尺寸与留白 |
| 总体方案 | 浏览器本地仓储（IndexedDB + Repository/mock API）+ HTML/CSS `@media print` |

### 1.4 非目标（MVP 不做）

- 真实后端 HTTP、MySQL、登录权限
- 打印模板可视化配置、导入/导出备份 UI
- 对接 `stock-center` / `stock-manage`
- 销售单状态机（草稿/已出库等）
- PDF 库出票、针式齿孔/复写联纹理还原
- 打入现有 Mac/Windows 一键分发包（Makefile 可后续加可选 target）

---

## 2. 架构

### 2.1 总体结构

```
浏览器（sales-front）
   │
   ├─ pages（列表 / 编辑 / 预览打印）
   ├─ salesOrderApi（外形像 HTTP）
   │       │
   │       ▼
   │  SalesOrderRepository → IndexedDB
   │
   └─ PrintProfile（v1 常量配置）
           │
           ▼
      OutboundSlip（出库单 DOM + print CSS）
```

与现有 `stock-*` **无运行时依赖**；仅同仓、同技术栈（Vite + React + TypeScript）。

### 2.2 持久化

| 项 | 选择 |
|----|------|
| 存储 | IndexedDB（建议 Dexie 薄封装） |
| 访问 | `salesOrderApi` → Repository，禁止页面直接操作 DB |
| 风险 | 清除站点数据会丢；MVP 不做导出，文档中注明 |
| 种子 | 可选 1～2 条 mock 销售单，便于首屏演示 |

### 2.3 以后换真 API

只替换 `salesOrderApi.ts`（或 Repository 实现）；页面与 `OutboundSlip` 不动。`PrintProfile` 再升级为可编辑配置即可。

---

## 3. 数据模型

### 3.1 销售单（持久化）

```ts
SalesOrder {
  id: string              // 本地 uuid
  orderNo: string         // 单据号，如 00150262
  orderDate: string       // ISO 日期；展示「YYYY年M月D日」
  customerName: string    // 客户
  warehouseName: string   // 仓库
  deliveryType: string    // 出库类型，如「提货」
  remark?: string         // 单据备注（表尾备注区）
  lines: SalesOrderLine[]
  totalQuantity: number   // Σ quantity（写入时重算）
  totalAmount: number     // Σ amount（写入时重算）
  createdAt: string
  updatedAt: string
}

SalesOrderLine {
  id: string
  materialName: string    // 物资
  spec: string            // 规格型号
  unit: string            // 单位
  quantity: number
  unitPrice: number
  amount: number          // 默认 quantity × unitPrice，允许覆盖
  lineRemark?: string     // 行备注
}
```

### 3.2 打印档案（v1 常量，不进销售单）

```ts
PrintProfile {
  companyName: string
  titleSuffix: string     // 如「出库单」；标题 = companyName + titleSuffix
  bankCardNo?: string     // 空则打印不显示该行
  bankCardHolder?: string
  address: string
  phone: string
  qualityNote: string
  paperWidthMm: number    // 初值约 190，对照样张微调
  paperHeightMm: number   // 初值约 140，对照样张微调
  tableMinRows: number    // 初值 10，不足补空行
}
```

### 3.3 字段归属

| 属于销售单 | 属于 PrintProfile |
|------------|-------------------|
| 日期、单据号、客户、仓库、出库类型 | 公司名、标题后缀「出库单」 |
| 明细七列、数量合计、金额合计 | 农行卡号、持卡人 |
| 单据备注；提货人签名区 v1 留空位即可 | 地址、电话、质量声明 |
| | 纸张 mm、最少表格行数 |

### 3.4 Repository / API 外形

| 动作 | 函数 | 说明 |
|------|------|------|
| 列表 | `listSalesOrders()` | 按 `updatedAt` 降序 |
| 详情 | `getSalesOrder(id)` | 不存在则明确错误 |
| 新建 | `createSalesOrder(input)` | 生成 id、时间戳、重算合计 |
| 更新 | `updateSalesOrder(id, input)` | 重算合计、更新 `updatedAt` |
| 删除 | `removeSalesOrder(id)` | MVP 可选暴露 |

非法输入（无客户、无明细行、数量/单价为负）在 API/表单层拒绝。

---

## 4. 页面与流程

```
/orders                 列表
/orders/new             新建
/orders/:id             编辑
/orders/:id/print       预览打印
```

### 4.1 列表页

- 列：单据号、日期、客户、仓库、出库类型、数量合计、金额合计、操作
- 操作：编辑、预览打印；删除带确认（可选）
- 空状态引导新建；可加载种子数据

### 4.2 编辑页

- 表头字段 + 明细增删行
- 改数量/单价默认重算行金额与单据合计
- 新建时本地生成单据号（可改）
- 校验：客户必填、至少一行、数量/单价 ≥ 0
- 保存后回列表（或提供「保存并预览」快捷，非必须）

### 4.3 预览打印页

- 只读渲染：`SalesOrder` + `PrintProfile` → `OutboundSlip`
- 明细行数 < `tableMinRows` 时补空行
- 工具条（返回、打印）在 `@media print` 中隐藏
- 数量/金额格式贴近样张（如 `784.00`、`5,409.60`）
- 打印：`window.print()`

---

## 5. 打印尺寸与还原

### 5.1 策略

- 屏幕：按 mm 比例显示票据画布
- 打印：同一套 DOM；物理单位 mm；隐藏非票据 UI
- 不引入 PDF 库

### 5.2 版式优先级

1. 结构：标题 → 表头字段 → 七列表格 → 合计 → 地址电话/备注/提货人 → 质量声明  
2. 对齐、字号、列宽比例贴近样张  
3. 细线表格；可用深蓝灰模拟针打字色；不强制模拟橙色底纸（底纸来自打印纸）  
4. 不还原：齿孔、复写联、针打纹理  

### 5.3 用户打印提示（实现时可放预览页文案）

- 关闭浏览器页眉页脚  
- 边距选默认或最小  
- 若系统按 A4 出纸：画布靠上/居中，周围留白可接受  

### 5.4 验收标准

- 预览字段位置与样张一致（允许 ±几 mm）  
- 打印预览内表格与底注不被裁切  
- 空行数量接近样张  

---

## 6. 目录与技术选型

### 6.1 技术

| 项 | 选择 |
|----|------|
| 框架 | Vite + React + TypeScript |
| 路由 | React Router |
| UI | 轻量组件库（Ant Design 或与 stock-front 同款） |
| 持久化 | IndexedDB + Dexie（或等价薄封装） |
| 打印 | `OutboundSlip` + `@media print` |

### 6.2 目录草案

```
apps/sales-front/
  package.json
  vite.config.ts
  index.html
  src/
    main.tsx
    App.tsx
    routes.tsx
    types/
      salesOrder.ts
      printProfile.ts
    config/
      printProfile.ts
      seed.ts
    storage/
      db.ts
      salesOrderRepository.ts
    services/
      salesOrderApi.ts
    pages/
      OrderListPage.tsx
      OrderEditPage.tsx
      OrderPrintPage.tsx
    components/
      OrderForm.tsx
      OrderLinesTable.tsx
      print/
        OutboundSlip.tsx
        outboundSlip.css
    utils/
      money.ts
      orderNo.ts
```

### 6.3 仓库集成

- 根 README / Makefile：可选补充 `dev-sales-front` / `build-sales-front`；不强制进入现有 pack 流程  
- 样张参考图可放在 `apps/sales-front/docs/` 或设计附件说明中，便于对照调样式  

---

## 7. 风险与后续

| 风险 | 缓解 |
|------|------|
| 清浏览器数据丢单 | 文档说明；后续可加 JSON 导出 |
| 不同打印机边距不一致 | PrintProfile 可调 mm；预览页提示 |
| 样张尺寸初值不准 | 实现后对照实物/扫描件微调 `paperWidthMm` / `paperHeightMm` |

**后续可选**：PrintProfile 配置页、导入导出、对接库存出库、打进分发包。

---

## 8. 参考样张字段映射（验收用）

| 样张文案 | 数据来源 |
|----------|----------|
| 标题公司名+出库单 | `PrintProfile.companyName` + `titleSuffix` |
| 日期 | `SalesOrder.orderDate` |
| 单据号 | `SalesOrder.orderNo` |
| 客户 | `SalesOrder.customerName` |
| 仓库 | `SalesOrder.warehouseName` |
| 出库类型 | `SalesOrder.deliveryType` |
| 农行卡号 / 持卡人 | `PrintProfile.bankCardNo` / `bankCardHolder` |
| 物资…备注列 | `SalesOrderLine.*` |
| 数量合计 / 合计 | `totalQuantity` / `totalAmount` |
| 地址 / 电话 / 质量备注 | `PrintProfile.*` |
| 备注 / 提货人 | `remark` + 签名空位 |
