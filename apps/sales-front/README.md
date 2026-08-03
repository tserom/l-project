# sales-front

销售单录单 + 出库单样式预览打印（本地 IndexedDB，无后端）。

## 开发

```bash
cd apps/sales-front && pnpm install && pnpm dev
```

打开 http://localhost:5175

或在仓库根目录：`make dev-sales-front`

## 发给别人（Windows 运行包）

在有 Node / pnpm 的机器上打包：

```bash
make pack-sales-front-windows
```

生成：`dist/sales-front-windows.zip`

对方：解压 → 双击 `start-sales.bat`（无需安装 Node）。说明见包内 `README.txt`。

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

领域设计摘要见 [`docs/domain-sales.md`](../../docs/domain-sales.md)；接口与 IndexedDB 见 [`docs/api-sales-storage.md`](../../docs/api-sales-storage.md)。

## 打印

预览页点击「打印」调用浏览器打印。建议：关闭页眉页脚；边距选默认或最小。纸张 mm 可在 `src/config/printProfile.ts` 微调。

## 注意

- 数据存在浏览器 IndexedDB；清除站点数据会丢失
- 与 stock-* 无运行时依赖
- 需要 Node ≥ 20（建议 `nvm use 20`）与 pnpm
- 测试：`pnpm test`
