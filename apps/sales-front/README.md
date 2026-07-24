# sales-front

销售单录单 + 出库单样式预览打印（本地 IndexedDB，无后端）。

## 开发

```bash
cd apps/sales-front && pnpm install && pnpm dev
```

打开 http://localhost:5175

或在仓库根目录：`make dev-sales-front`

## 注意

- 数据存在浏览器 IndexedDB；清除站点数据会丢失
- 与 stock-* 无运行时依赖
- 需要 Node ≥ 20（建议 `nvm use 20`）与 pnpm
