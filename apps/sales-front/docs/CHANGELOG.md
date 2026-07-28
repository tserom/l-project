# sales-front 修改记录

按时间倒序记录面向业务的变更，便于留档与发包对照。

---

## 2026-07-28 — 列表查询条件

**类型**：功能  
**分支**：`feature/sales-front`

### 变更说明

销售单列表页增加查询区（本地 IndexedDB 数据前端过滤，无后端）：

| 条件 | 规则 |
|------|------|
| 单据号 | 多值（回车 / 中英文逗号 / 空格分隔），与任一单号**精确**匹配 |
| 日期 | 范围选择，起止日期均含当日 |
| 客户 | 客户名称**包含**匹配（模糊） |

提供「查询」「重置」；无命中时提示「无符合条件的销售单」。

### 涉及文件

- `src/pages/OrderListPage.tsx` — 查询表单 + 过滤后的表格数据源
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
