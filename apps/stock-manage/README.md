# stock-manage

面向业务的库存管理服务，对外提供业务 API，通过 HTTP 调用 **stock-center** 完成库存数据读写，并在本地 MySQL 记录业务操作日志。

## 技术栈

- Gin — HTTP 框架
- GORM — ORM（业务库，如操作审计）
- MySQL — 持久化
- HTTP Client — 调用 stock-center

## 分层

```
cmd/server              入口
internal/config         配置
internal/database       业务库连接与迁移
internal/model          业务模型
internal/repository     业务库访问
internal/client/stockcenter  stock-center 上游客户端
internal/service        业务编排
internal/handler        HTTP 处理
internal/router         路由注册
pkg/response            统一响应
```

## 架构关系

```
前端 / 其它服务
      │
      ▼
stock-manage (8082)  ──HTTP──▶  stock-center (8081) ──▶ MySQL (stock_center)
      │
      └──▶ MySQL (stock_manage)  业务审计表
```

## 本地开发

先启动 stock-center，再启动本服务：

```bash
# 终端 1
cd apps/stock-center && make run

# 终端 2
cd apps/stock-manage
cp .env.example .env
export $(grep -v '^#' .env | xargs)
make tidy
make run
```

默认端口：**8082**

健康检查：`GET /health`

## API（示例）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/inventory` | 分页查询库存（转发 stock-center） |
| GET | `/api/v1/inventory/query?sku=&warehouse=` | 按 SKU 查询库存 |
| POST | `/api/v1/inventory/inbound` | 业务入库（写 stock-center + 记审计日志） |
