# sales-front

销售单录单 + 出库单样式预览打印 + 开票（前端）。

持久化由 **`sales-manage`（Go + MySQL，:8083）** 提供，不再使用 IndexedDB。

## 开发

先起后端与 MySQL（库名 `sales_manage`），再起前端：

```bash
# 终端 1
make dev-sales-manage
# 或：cd apps/sales-manage && cp .env.example .env && export $(grep -v '^#' .env | xargs) && make run

# 终端 2
cd apps/sales-front && pnpm install && pnpm dev
```

打开 http://localhost:5175  

Vite 已将 `/api` 代理到 `http://127.0.0.1:8083`。

或根目录：`make dev-sales-front`（仍需先起 `sales-manage`）。

## 发给别人（Windows 运行包）

`make pack-sales-front-windows` 仍打静态包；对方机器还需可访问的 `sales-manage` + MySQL（或后续再做一体打包）。说明见包内 `README.txt`（待同步时改）。

## 修改记录

见 [`docs/CHANGELOG.md`](docs/CHANGELOG.md)。

## 路由

| 路径 | 说明 |
|------|------|
| `/orders` | 销售单列表 |
| `/orders/new` | 新建 |
| `/orders/:id` | 编辑 |
| `/orders/:id/print` | 预览打印 |
| `/invoices` | 开票单列表 |
| `/invoices/new` | 新建开票单 |
| `/invoices/:id` | 开票单详情 |

领域：[`docs/domain-sales.md`](../../docs/domain-sales.md)；HTTP：[`docs/api-sales-storage.md`](../../docs/api-sales-storage.md)；后端：`apps/sales-manage/README.md`。

## 打印

预览页点击「打印」调用浏览器打印。建议：关闭页眉页脚；边距选默认或最小。纸张 mm 可在 `src/config/printProfile.ts` 微调。

## 注意

- 真源为 MySQL（`sales_manage`）；与 stock-* **无**业务打通（stock-manage `/sales-orders` 是另一套）
- IndexedDB → MySQL 历史数据导入为第二期
- 需要 Node ≥ 20（建议 `nvm use 20`）与 pnpm
- 测试：`pnpm test`
