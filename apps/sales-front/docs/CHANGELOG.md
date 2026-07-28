# sales-front 修改记录

按时间倒序记录面向业务的变更，便于留档与发包对照。

---

## 2026-07-28 — 开票单（明细需开票 + 汇总开票）

**类型**：功能  
**分支**：`feature/sales-front`  
**设计**：`docs/superpowers/specs/2026-07-28-invoice-doc-design.md`

### 变更说明

- 销售明细增加「需开票」勾选（默认否）；已开票显示标签，不可删行
- 新增开票单：可选客户 / 日期 / 单据号加载待开票明细，可删行后保存
- 保存后明细挂开票单；作废后明细重新可开
- 开票明细行末展示来源销售单单号、日期、客户、仓库、出库类型

### 涉及要点

- IndexedDB `invoiceDocs`（schema v2）
- 顶栏导航：销售单 / 开票单

### 验证

- `pnpm test` 通过（含 create/void 开票单用例）

---

## 2026-07-28 — 列表查询条件与金额合计

**类型**：功能  
**分支**：`feature/sales-front`

### 变更说明

销售单列表页增加查询区（本地 IndexedDB 数据前端过滤，无后端）：

| 条件 | 规则 |
|------|------|
| 单据号 | 多值（回车 / 中英文逗号 / 空格分隔），与任一单号**精确**匹配 |
| 日期 | 范围选择，起止日期均含当日 |
| 客户 | 客户名称**包含**匹配（模糊） |

表格底部增加当前筛选结果的**金额合计**。

### 涉及文件

- `src/pages/OrderListPage.tsx` — 查询表单 + 过滤后的表格数据源 + 合计行
- `src/utils/filterSalesOrders.ts` — 过滤与多值单号解析
- `src/utils/filterSalesOrders.test.ts` — 单测

### 验证

- `pnpm test`（含 filter 用例）通过

---

## 2026-07-24 — MVP 与 Windows 分享包

**类型**：功能 / 修复  
**分支**：`feature/sales-front`

### 摘要

- 新建 `apps/sales-front`：销售单 CRUD、IndexedDB、出库单样式预览打印
- Windows 运行包：`make pack-sales-front-windows` → `dist/sales-front-windows.zip`
- 修复：明细行拼音 IME 打断；bat UTF-8 中文导致 cmd 乱码；扁平包结构（`index.html` 与 bat 同级）

### 设计 / 计划

- `docs/superpowers/specs/2026-07-24-sales-order-app-design.md`
- `docs/superpowers/plans/2026-07-24-sales-order-app.md`
