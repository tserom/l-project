# sales-front

销售单录单 + 出库单样式预览打印（本地 IndexedDB，无后端）。

## 开发

```bash
cd apps/sales-front && pnpm install && pnpm dev
```

打开 http://localhost:5175

或在仓库根目录：`make dev-sales-front`

## 路由

| 路径 | 说明 |
|------|------|
| `/orders` | 销售单列表 |
| `/orders/new` | 新建 |
| `/orders/:id` | 编辑 |
| `/orders/:id/print` | 预览打印 |

## 打印

预览页点击「打印」调用浏览器打印。建议：关闭页眉页脚；边距选默认或最小。纸张 mm 可在 `src/config/printProfile.ts` 微调。

## 注意

- 数据存在浏览器 IndexedDB；清除站点数据会丢失
- 与 stock-* 无运行时依赖
- 需要 Node ≥ 20（建议 `nvm use 20`）与 pnpm
- 测试：`pnpm test`
