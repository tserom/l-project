# stock-center

库存数据中心服务，**直接操作 MySQL**，为上层业务服务提供库存数据的读写能力。

## 技术栈

- Gin — HTTP 框架
- GORM — ORM
- MySQL — 持久化

## 分层

```
cmd/server          入口
internal/config     配置
internal/database   数据库连接与迁移
internal/model      数据模型
internal/repository 直接数据库访问（L1 数据层）
internal/service    领域服务
internal/handler    HTTP 处理
internal/router     路由注册
pkg/response        统一响应
```

## 本地开发

```bash
cd apps/stock-center
cp .env.example .env   # 按需修改数据库连接
export $(grep -v '^#' .env | xargs)
make tidy
make run
```

默认端口：**8081**

健康检查：`GET /health`

## API（示例）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/stocks` | 分页列表 |
| GET | `/api/v1/stocks/:id` | 按 ID 查询 |
| GET | `/api/v1/stocks/by-sku?sku=&warehouse=` | 按 SKU + 仓库查询 |
| POST | `/api/v1/stocks` | 创建库存记录 |
| PUT | `/api/v1/stocks/:id/quantity` | 调整库存数量 |
