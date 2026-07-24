# 销售单 App（sales-front）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在本仓新建独立前端 `apps/sales-front`，支持销售单本地 CRUD、IndexedDB 持久化，以及按连续纸出库单样式预览并浏览器打印。

**Architecture:** 页面只调 `salesOrderApi`；Api 委托 `SalesOrderRepository` 读写 IndexedDB（Dexie）。打印用同一套 `OutboundSlip` DOM + `@media print`（mm 画布）。`PrintProfile` v1 为常量，与销售单字段分离。

**Tech Stack:** Vite 6+、React 19、TypeScript、React Router 7、Ant Design 5、Dexie 4、Vitest、pnpm（与仓库其它前端一致用 pnpm）

**关联文档：**

- 设计 spec：`docs/superpowers/specs/2026-07-24-sales-order-app-design.md`
- 样张参考：对话中出库单照片（实现打印时对照；可选复制到 `apps/sales-front/docs/sample-slip.png`）

## Global Constraints

- 业务实体是**销售单**；打印标题为公司名 +「出库单」样式，不是另建出库单实体
- 独立 App：`apps/sales-front`，**不**依赖 `stock-center` / `stock-manage` / `stock-front`
- 无真实 HTTP 后端、无 MySQL；持久化仅 IndexedDB
- 页面**禁止**直接操作 Dexie；必须经 `salesOrderApi`
- MVP **不做**：模板配置 UI、导入导出、登录、PDF 库、打进现有 pack 脚本
- UI 文案中文；金额展示两位小数；合计金额可用千分位（如 `5,409.60`）
- 注意：现有 `stock-front` 是 **Vue**；本 App 按已批准设计使用 **React**，不要改成 Vue「对齐」

---

## 文件结构总览

| 路径 | 职责 |
|------|------|
| `apps/sales-front/package.json` | 依赖与 scripts（dev/build/test） |
| `apps/sales-front/vite.config.ts` | Vite + React + Vitest |
| `apps/sales-front/index.html` | 入口 HTML |
| `apps/sales-front/src/main.tsx` | React 挂载 + Ant Design ConfigProvider（zhCN） |
| `apps/sales-front/src/App.tsx` | 布局壳（可选简单 Header） |
| `apps/sales-front/src/routes.tsx` | 路由表 |
| `apps/sales-front/src/types/salesOrder.ts` | `SalesOrder` / `SalesOrderLine` / `SalesOrderInput` |
| `apps/sales-front/src/types/printProfile.ts` | `PrintProfile` |
| `apps/sales-front/src/utils/money.ts` | 行金额、合计、格式化 |
| `apps/sales-front/src/utils/orderNo.ts` | 单据号生成 |
| `apps/sales-front/src/utils/dateFormat.ts` | ISO → `YYYY年M月D日` |
| `apps/sales-front/src/config/printProfile.ts` | v1 打印常量（样张公司信息） |
| `apps/sales-front/src/config/seed.ts` | 可选种子销售单 |
| `apps/sales-front/src/storage/db.ts` | Dexie 数据库定义 |
| `apps/sales-front/src/storage/salesOrderRepository.ts` | IndexedDB CRUD |
| `apps/sales-front/src/services/salesOrderApi.ts` | 校验 + 委托 Repository；入口索引注释 |
| `apps/sales-front/src/pages/OrderListPage.tsx` | 列表 |
| `apps/sales-front/src/pages/OrderEditPage.tsx` | 新建/编辑 |
| `apps/sales-front/src/pages/OrderPrintPage.tsx` | 预览 + 打印 |
| `apps/sales-front/src/components/OrderForm.tsx` | 表头表单 |
| `apps/sales-front/src/components/OrderLinesTable.tsx` | 明细编辑表 |
| `apps/sales-front/src/components/print/OutboundSlip.tsx` | 出库单版式 |
| `apps/sales-front/src/components/print/outboundSlip.css` | 屏显 + `@media print` |
| `apps/sales-front/src/utils/*.test.ts` 等 | Vitest 单测 |
| `apps/sales-front/README.md` | 启动、持久化风险说明 |
| `Makefile`（可选） | `dev-sales-front` / `build-sales-front` |

---

### Task 1: 脚手架 `apps/sales-front`

**Files:**
- Create: `apps/sales-front/`（Vite React-TS 工程）
- Create: `apps/sales-front/README.md`
- Modify: `Makefile`（追加两个 phony target）

**Interfaces:**
- Produces: 可 `pnpm dev` 的空 React App；Vitest 可跑空套件

- [ ] **Step 1: 用 Vite 创建工程**

```bash
cd /Users/lingqirui/Desktop/l-project
pnpm create vite apps/sales-front --template react-ts
cd apps/sales-front
pnpm install
pnpm add react-router-dom antd dayjs dexie
pnpm add -D vitest jsdom @testing-library/react @testing-library/jest-dom @testing-library/user-event fake-indexeddb
```

若 `create vite` 交互失败，手动写入等价 `package.json` / `vite.config.ts` / `tsconfig*` / `index.html` / `src/main.tsx`。

- [ ] **Step 2: 配置 Vitest 与路径别名**

`apps/sales-front/vite.config.ts`：

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, 'src') },
  },
  server: { port: 5175 },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    globals: true,
  },
})
```

`apps/sales-front/src/test/setup.ts`：

```ts
import '@testing-library/jest-dom/vitest'
import 'fake-indexeddb/auto'
```

`package.json` scripts 增加：

```json
"test": "vitest run",
"test:watch": "vitest",
"dev": "vite --port 5175"
```

确保 `tsconfig.app.json` 含 `"paths": { "@/*": ["./src/*"] }`（与 Vite alias 一致）。

- [ ] **Step 3: 写 README 与 Makefile 入口**

`apps/sales-front/README.md` 至少包含：开发命令、IndexedDB 清站点会丢数据、与 stock-* 无运行时依赖。

根 `Makefile` 追加：

```makefile
.PHONY: build-front build-center build-manage build-all pack-mac pack-windows dev-sales-front build-sales-front

dev-sales-front:
	cd apps/sales-front && pnpm install && pnpm dev

build-sales-front:
	cd apps/sales-front && pnpm install && pnpm build
```

- [ ] **Step 4: 验证脚手架**

```bash
cd apps/sales-front && pnpm test && pnpm build
```

Expected: Vitest 无用例或 0 失败；`build` 成功。

- [ ] **Step 5: Commit**

```bash
git add apps/sales-front Makefile
git commit -m "$(cat <<'EOF'
chore: scaffold apps/sales-front (Vite React TS)

EOF
)"
```

---

### Task 2: 类型与纯函数工具（TDD）

**Files:**
- Create: `apps/sales-front/src/types/salesOrder.ts`
- Create: `apps/sales-front/src/types/printProfile.ts`
- Create: `apps/sales-front/src/utils/money.ts`
- Create: `apps/sales-front/src/utils/orderNo.ts`
- Create: `apps/sales-front/src/utils/dateFormat.ts`
- Create: `apps/sales-front/src/utils/money.test.ts`
- Create: `apps/sales-front/src/utils/orderNo.test.ts`
- Create: `apps/sales-front/src/utils/dateFormat.test.ts`

**Interfaces:**
- Produces:
  - `SalesOrder`, `SalesOrderLine`, `SalesOrderInput`（创建/更新载荷；合计与时间戳由 Api 写）
  - `calcLineAmount(qty, price): number`
  - `sumQuantities(lines): number` / `sumAmounts(lines): number`
  - `formatQuantity(n): string` → `"784.00"`
  - `formatAmount(n): string` → `"5,409.60"`
  - `formatOrderDate(isoDate: string): string` → `"2026年7月22日"`
  - `generateOrderNo(now?: Date): string`

- [ ] **Step 1: 写失败测试**

`apps/sales-front/src/utils/money.test.ts`：

```ts
import { describe, expect, it } from 'vitest'
import { calcLineAmount, formatAmount, formatQuantity, sumAmounts, sumQuantities } from './money'

describe('money', () => {
  it('calcLineAmount rounds to 2 decimals', () => {
    expect(calcLineAmount(784, 6.9)).toBe(5409.6)
  })

  it('sums lines', () => {
    const lines = [
      { quantity: 784, amount: 5409.6 },
      { quantity: 10, amount: 69 },
    ]
    expect(sumQuantities(lines)).toBe(794)
    expect(sumAmounts(lines)).toBe(5478.6)
  })

  it('formats like sample slip', () => {
    expect(formatQuantity(784)).toBe('784.00')
    expect(formatAmount(5409.6)).toBe('5,409.60')
  })
})
```

`dateFormat.test.ts`：`formatOrderDate('2026-07-22')` → `'2026年7月22日'`。

`orderNo.test.ts`：`generateOrderNo` 返回非空字符串，长度 ≥ 8。

- [ ] **Step 2: 跑测试确认失败**

```bash
cd apps/sales-front && pnpm test
```

Expected: FAIL（模块不存在或导出缺失）。

- [ ] **Step 3: 实现类型与工具**

`types/salesOrder.ts`：

```ts
export interface SalesOrderLine {
  id: string
  materialName: string
  spec: string
  unit: string
  quantity: number
  unitPrice: number
  amount: number
  lineRemark?: string
}

export interface SalesOrder {
  id: string
  orderNo: string
  orderDate: string
  customerName: string
  warehouseName: string
  deliveryType: string
  remark?: string
  lines: SalesOrderLine[]
  totalQuantity: number
  totalAmount: number
  createdAt: string
  updatedAt: string
}

/** 创建/更新时由调用方提供的字段（合计与时间戳由 Api 写入） */
export type SalesOrderInput = Omit<
  SalesOrder,
  'id' | 'totalQuantity' | 'totalAmount' | 'createdAt' | 'updatedAt'
> & { id?: string }
```

`types/printProfile.ts`：

```ts
export interface PrintProfile {
  companyName: string
  titleSuffix: string
  bankCardNo?: string
  bankCardHolder?: string
  address: string
  phone: string
  qualityNote: string
  paperWidthMm: number
  paperHeightMm: number
  tableMinRows: number
}
```

`money.ts`：用 `Math.round(n * 100) / 100`；`formatAmount` 用 `toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })`。

`dateFormat.ts`：解析 `YYYY-MM-DD`，月日不补零。

`orderNo.ts`：例如 `` `${yyyy}${mm}${dd}${String(now.getTime() % 10000).padStart(4, '0')}` ``。

- [ ] **Step 4: 跑测试确认通过**

```bash
cd apps/sales-front && pnpm test
```

Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add apps/sales-front/src/types apps/sales-front/src/utils
git commit -m "$(cat <<'EOF'
feat(sales-front): add sales order types and format helpers

EOF
)"
```

---

### Task 3: PrintProfile 常量与种子数据

**Files:**
- Create: `apps/sales-front/src/config/printProfile.ts`
- Create: `apps/sales-front/src/config/seed.ts`

**Interfaces:**
- Produces: `export const defaultPrintProfile: PrintProfile`
- Produces: `export const seedSalesOrders: SalesOrder[]`（1 条对齐样张）

- [ ] **Step 1: 写入 PrintProfile（对照样张）**

```ts
import type { PrintProfile } from '@/types/printProfile'

export const defaultPrintProfile: PrintProfile = {
  companyName: '泰州市金阳金属制品有限公司',
  titleSuffix: '出库单',
  bankCardNo: '6230523420034407572',
  bankCardHolder: '刘敏',
  address: '兴化市戴南镇裴马工业区',
  phone: '13705265020',
  qualityNote:
    '备注：如出现质量问题请于收到货7日内以书面形式提出质量异议，逾期则视为合格产品。35mm以上棒料严禁下料机（冲床）下料，出现裂纹后果自负。',
  paperWidthMm: 190,
  paperHeightMm: 140,
  tableMinRows: 10,
}
```

- [ ] **Step 2: 写入种子销售单**

`seed.ts` 一条 `SalesOrder`：单据号 `00150262`、客户 `884周村 马俊生`、仓库 `01金阳仓库`、出库类型 `提货`、一行 `002024-2Cr13黑棒` / `Φ 32` / `kg` / `784` / `6.9` / `5409.6`。固定 `id: 'seed-1'`，便于幂等。

- [ ] **Step 3: Commit**

```bash
git add apps/sales-front/src/config
git commit -m "$(cat <<'EOF'
feat(sales-front): add print profile constants and seed order

EOF
)"
```

---

### Task 4: Dexie Repository + salesOrderApi（TDD）

**Files:**
- Create: `apps/sales-front/src/storage/db.ts`
- Create: `apps/sales-front/src/storage/salesOrderRepository.ts`
- Create: `apps/sales-front/src/services/salesOrderApi.ts`
- Create: `apps/sales-front/src/services/salesOrderApi.test.ts`

**Interfaces:**
- Consumes: `SalesOrder`, `SalesOrderInput`, money helpers
- Produces（Api，文件顶部注释 call-map）：

```ts
listSalesOrders(): Promise<SalesOrder[]>
getSalesOrder(id: string): Promise<SalesOrder>
createSalesOrder(input: SalesOrderInput): Promise<SalesOrder>
updateSalesOrder(id: string, input: SalesOrderInput): Promise<SalesOrder>
removeSalesOrder(id: string): Promise<void>
ensureSeedData(seeds: SalesOrder[]): Promise<void>  // 仅当库空时写入
```

校验与写入规则：
- `customerName` trim 后非空；`lines.length >= 1`；每行 `quantity >= 0`、`unitPrice >= 0`
- 保存时用 `calcLineAmount` **重算**每行 `amount`（避免脏数据）
- `totalQuantity` / `totalAmount` 由 Api 重算
- `create`：`id = crypto.randomUUID()`，时间戳 ISO
- `update`：保留 `createdAt`，刷新 `updatedAt`
- `get` 不存在抛 `Error('销售单不存在')`
- `list` 按 `updatedAt` 降序

- [ ] **Step 1: 写失败测试**

`salesOrderApi.test.ts`：

```ts
import { beforeEach, describe, expect, it } from 'vitest'
import { db } from '@/storage/db'
import {
  createSalesOrder,
  getSalesOrder,
  listSalesOrders,
  removeSalesOrder,
  updateSalesOrder,
  ensureSeedData,
} from './salesOrderApi'
import { seedSalesOrders } from '@/config/seed'

const validInput = {
  orderNo: '00150262',
  orderDate: '2026-07-22',
  customerName: '884周村 马俊生',
  warehouseName: '01金阳仓库',
  deliveryType: '提货',
  lines: [
    {
      id: 'line-1',
      materialName: '002024-2Cr13黑棒',
      spec: 'Φ 32',
      unit: 'kg',
      quantity: 784,
      unitPrice: 6.9,
      amount: 0,
    },
  ],
}

beforeEach(async () => {
  await db.delete()
  await db.open()
})

describe('salesOrderApi', () => {
  it('creates, lists, gets, updates, removes', async () => {
    const created = await createSalesOrder(validInput)
    expect(created.totalAmount).toBe(5409.6)
    expect(created.totalQuantity).toBe(784)
    expect(await listSalesOrders()).toHaveLength(1)
    const got = await getSalesOrder(created.id)
    expect(got.orderNo).toBe('00150262')
    await updateSalesOrder(created.id, { ...validInput, customerName: '新客户' })
    expect((await getSalesOrder(created.id)).customerName).toBe('新客户')
    await removeSalesOrder(created.id)
    expect(await listSalesOrders()).toHaveLength(0)
  })

  it('rejects empty customer', async () => {
    await expect(
      createSalesOrder({ ...validInput, customerName: '  ' }),
    ).rejects.toThrow(/客户/)
  })

  it('ensureSeedData only when empty', async () => {
    await ensureSeedData(seedSalesOrders)
    await ensureSeedData(seedSalesOrders)
    expect(await listSalesOrders()).toHaveLength(seedSalesOrders.length)
  })
})
```

- [ ] **Step 2: 跑测试确认失败**

```bash
cd apps/sales-front && pnpm test src/services/salesOrderApi.test.ts
```

Expected: FAIL。

- [ ] **Step 3: 实现 db / repository / api**

`db.ts`：

```ts
import Dexie, { type EntityTable } from 'dexie'
import type { SalesOrder } from '@/types/salesOrder'

export class SalesFrontDB extends Dexie {
  salesOrders!: EntityTable<SalesOrder, 'id'>

  constructor() {
    super('sales-front')
    this.version(1).stores({
      salesOrders: 'id, orderNo, updatedAt',
    })
  }
}

export const db = new SalesFrontDB()
```

Repository：`put` / `get` / `toArray` / `delete`；Api 做校验与合计。文件顶部写 call-map 注释（动作 → 函数 → IndexedDB）。

- [ ] **Step 4: 跑测试确认通过**

```bash
cd apps/sales-front && pnpm test
```

Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add apps/sales-front/src/storage apps/sales-front/src/services
git commit -m "$(cat <<'EOF'
feat(sales-front): IndexedDB repository and salesOrderApi

EOF
)"
```

---

### Task 5: 路由壳与列表页

**Files:**
- Modify: `apps/sales-front/src/main.tsx`
- Create: `apps/sales-front/src/App.tsx`
- Create: `apps/sales-front/src/routes.tsx`
- Create: `apps/sales-front/src/pages/OrderListPage.tsx`
- Create: `apps/sales-front/src/pages/OrderEditPage.tsx`（占位文案）
- Create: `apps/sales-front/src/pages/OrderPrintPage.tsx`（占位文案）

**Interfaces:**
- Consumes: `listSalesOrders`, `removeSalesOrder`, `ensureSeedData`, `seedSalesOrders`, format helpers
- Produces: `/` → `/orders`；`/orders`；`/orders/new`；`/orders/:id`；`/orders/:id/print`

- [ ] **Step 1: 挂载 Router + Ant Design zhCN**

`main.tsx`：`ConfigProvider locale={zhCN}` + Router + `App`。

`routes.tsx`：注册上述路径。

- [ ] **Step 2: 实现 OrderListPage**

- `useEffect`：先 `ensureSeedData(seedSalesOrders)`，再 `listSalesOrders()`
- `Table` 列：单据号、日期（`formatOrderDate`）、客户、仓库、出库类型、数量合计、金额合计、操作
- 操作：编辑、预览打印；删除 `Modal.confirm` → `removeSalesOrder` → 刷新
- 「新建销售单」→ `/orders/new`；空态 `Empty`

- [ ] **Step 3: 手工验证**

```bash
cd apps/sales-front && pnpm dev
```

Expected: 列表可见种子单；刷新仍在；删除后消失；其它路由进占位页。

- [ ] **Step 4: Commit**

```bash
git add apps/sales-front/src
git commit -m "$(cat <<'EOF'
feat(sales-front): order list page with seed and routes

EOF
)"
```

---

### Task 6: 编辑表单（OrderForm + OrderLinesTable + OrderEditPage）

**Files:**
- Create: `apps/sales-front/src/components/OrderForm.tsx`
- Create: `apps/sales-front/src/components/OrderLinesTable.tsx`
- Modify: `apps/sales-front/src/pages/OrderEditPage.tsx`

**Interfaces:**
- Consumes: `getSalesOrder`, `createSalesOrder`, `updateSalesOrder`, `generateOrderNo`, `calcLineAmount`
- Produces: 可保存的新建/编辑页；成功后 `navigate('/orders')`

- [ ] **Step 1: OrderLinesTable**

```ts
type Props = {
  value: SalesOrderLine[]
  onChange: (lines: SalesOrderLine[]) => void
}
```

- 列：物资、规格型号、单位、数量、单价、金额（展示重算）、行备注、删除
- 「添加行」：`id: crypto.randomUUID()`，默认 `unit: 'kg'`
- 改数量/单价 → `amount = calcLineAmount(...)`
- 底部 `formatQuantity(sumQuantities)` / `formatAmount(sumAmounts)`

- [ ] **Step 2: OrderForm**

表头：`orderDate`（DatePicker → `YYYY-MM-DD`）、`orderNo`、`customerName`、`warehouseName`、`deliveryType`、`remark` + `OrderLinesTable`。

- [ ] **Step 3: OrderEditPage**

- 路由 `/orders/new` 为新建，`/orders/:id` 为编辑
- 新建：`orderNo = generateOrderNo()`，日期今天，`deliveryType = '提货'`，一行空明细
- 编辑：`getSalesOrder(id)`；失败提示回列表
- 保存：`createSalesOrder` / `updateSalesOrder`；`message.success`；回列表
- 捕获 Api 中文错误并展示

- [ ] **Step 4: 手工验证**

新建 → 列表出现 → 编辑改客户 → 合计随明细变化。

- [ ] **Step 5: Commit**

```bash
git add apps/sales-front/src/components apps/sales-front/src/pages/OrderEditPage.tsx
git commit -m "$(cat <<'EOF'
feat(sales-front): sales order create and edit form

EOF
)"
```

---

### Task 7: OutboundSlip 打印组件（屏显版式）

**Files:**
- Create: `apps/sales-front/src/components/print/OutboundSlip.tsx`
- Create: `apps/sales-front/src/components/print/outboundSlip.css`
- Create: `apps/sales-front/src/components/print/padLines.ts`
- Create: `apps/sales-front/src/components/print/padLines.test.ts`

**Interfaces:**
- Consumes: `SalesOrder`, `PrintProfile`, format helpers
- Produces:

```ts
export function padLinesToMin(
  lines: SalesOrderLine[],
  minRows: number,
): Array<SalesOrderLine | null>

export function OutboundSlip(props: {
  order: SalesOrder
  profile: PrintProfile
}): JSX.Element
```

- [ ] **Step 1: padLines 测试与实现**

```ts
it('pads with null to min rows', () => {
  expect(padLinesToMin([], 3)).toEqual([null, null, null])
  expect(padLinesToMin([{ id: '1' } as SalesOrderLine], 2)).toHaveLength(2)
})
```

- [ ] **Step 2: OutboundSlip 结构（对照样张）**

根节点 `.outbound-slip`：

```css
.outbound-slip {
  width: var(--slip-width, 190mm);
  min-height: var(--slip-height, 140mm);
  color: #1a3a6b;
  font-family: "SimSun", "Songti SC", serif;
  box-sizing: border-box;
  padding: 4mm 6mm;
}
```

DOM 顺序：标题 → 日期/单据号 → 客户/仓库/出库类型/农行卡号（有则显示）→ 七列表格（`padLinesToMin`）→ 数量合计/合计 → 地址/电话/备注/提货人 → `qualityNote`。

数字列右对齐；空行空白 `<tr>`。用 CSS 变量传入 `paperWidthMm` / `paperHeightMm`。

- [ ] **Step 3: 跑 padLines 测试**

```bash
cd apps/sales-front && pnpm test src/components/print/padLines.test.ts
```

Expected: PASS。

- [ ] **Step 4: Commit**

```bash
git add apps/sales-front/src/components/print
git commit -m "$(cat <<'EOF'
feat(sales-front): OutboundSlip layout component

EOF
)"
```

---

### Task 8: 预览打印页 + @media print

**Files:**
- Modify: `apps/sales-front/src/pages/OrderPrintPage.tsx`
- Modify: `apps/sales-front/src/components/print/outboundSlip.css`

**Interfaces:**
- Consumes: `getSalesOrder`, `defaultPrintProfile`, `OutboundSlip`
- Produces: 工具条（返回、打印）+ 画布；打印隐藏工具条

- [ ] **Step 1: OrderPrintPage**

- 加载 `getSalesOrder(id)`；失败回列表
- 工具条 `className="no-print"`：返回、打印（`window.print()`）
- 提示：关闭页眉页脚；边距选默认或最小
- `<OutboundSlip order={order} profile={defaultPrintProfile} />`

- [ ] **Step 2: 打印 CSS**

```css
@media print {
  .no-print {
    display: none !important;
  }
  @page {
    size: auto;
    margin: 8mm;
  }
  body {
    margin: 0;
  }
  .outbound-slip {
    box-shadow: none;
  }
}
```

屏显可给画布浅阴影；打印去掉。

- [ ] **Step 3: 手工验收（对照样张 / spec §8）**

1. 种子单预览字段正确  
2. Chrome 打印预览表格与底注不被裁切  
3. 空行观感接近样张  
4. 偏差大时只调 `paperWidthMm` / `paperHeightMm` 与 CSS  

- [ ] **Step 4: Commit**

```bash
git add apps/sales-front/src/pages/OrderPrintPage.tsx apps/sales-front/src/components/print
git commit -m "$(cat <<'EOF'
feat(sales-front): print preview page with media print styles

EOF
)"
```

---

### Task 9: 收尾与回归

**Files:**
- Modify: `apps/sales-front/README.md`（路由、打印提示）
- Modify: `docs/superpowers/specs/2026-07-24-sales-order-app-design.md`（状态改为已批准并指向本 plan）

- [ ] **Step 1: 全量测试与构建**

```bash
cd apps/sales-front && pnpm test && pnpm build
```

Expected: 全部 PASS；`dist/` 生成成功。

- [ ] **Step 2: 端到端手工清单**

- [ ] 列表有种子；刷新仍在  
- [ ] 新建 → 编辑 → 删除  
- [ ] 预览字段正确；可调出打印对话框  
- [ ] 无控制台报错  

- [ ] **Step 3: Commit**

```bash
git add apps/sales-front/README.md docs/superpowers/specs/2026-07-24-sales-order-app-design.md
git commit -m "$(cat <<'EOF'
docs(sales-front): finish README and mark design approved for impl

EOF
)"
```

---

## Spec 覆盖自检

| Spec 要求 | Task |
|-----------|------|
| 独立 `apps/sales-front` | 1 |
| SalesOrder / Line / PrintProfile 分离 | 2, 3 |
| IndexedDB + Api 外形 | 4 |
| list/get/create/update/remove + 种子 | 4, 5 |
| 列表 / 编辑 / 打印三页 | 5, 6, 8 |
| 出库单版式 + mm + 空行 | 7, 8 |
| `@media print` / window.print | 8 |
| Makefile 可选入口 | 1 |
| 不做模板配置/导入导出/后端 | 全任务未包含 |

**类型一致性：** `SalesOrderInput`、Api 函数名在 Task 2/4 定义，Task 5–8 只消费这些名字；`defaultPrintProfile` 在 Task 3，Task 8 使用。

**占位符扫描：** 无 TBD/TODO；步骤含具体代码与命令。
