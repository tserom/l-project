# sales-manage

销售开票工具后端：承接 `sales-front` 的销售单与开票单持久化（MySQL）。

**不是** `stock-manage` 的库存域 `/api/v1/sales-orders`。

## 技术栈

- Gin + GORM + MySQL
- 默认端口 `8083`，库名 `sales_manage`

## 本地启动

```bash
# 先建库
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS sales_manage DEFAULT CHARSET utf8mb4;"

cp .env.example .env
export $(grep -v '^#' .env | xargs)
make run
# → http://127.0.0.1:8083/health
```

前端：`apps/sales-front` 通过 Vite 代理 `/api` → `8083`。

## 主要路由

| 方法 | 路径 |
|------|------|
| CRUD | `/api/v1/sale/orders` |
| 开票候选 | `GET /api/v1/sale/orders/invoice-candidates` |
| 开票 | `/api/v1/sale/invoice-docs` |
| 作废 | `POST /api/v1/sale/invoice-docs/:id/void` |

列表筛选：`qp-<field>-<operator>` + `page` / `pageSize`。

## 测试

```bash
make test
```

使用 sqlite 内存库，无需 MySQL。

## 说明

- IndexedDB → MySQL 导入脚本为第二期
- 与 stock-center / stock-manage 无运行时依赖
