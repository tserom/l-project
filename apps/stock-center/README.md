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

## 一键启动

先构建二进制（推荐）：

```bash
# 仓库根目录
make build-all
```

启动 MySQL 并初始化库（首次）：

```bash
mysql -u root -p < scripts/mysql/init-databases.sql
cp apps/stock-center/.env.example apps/stock-center/.env
cp apps/stock-manage/.env.example apps/stock-manage/.env
```

Mac / Linux：

```bash
chmod +x scripts/launcher/start.sh
./scripts/launcher/start.sh
```

Windows：

```bat
scripts\launcher\start.bat
```

启动器会依次拉起 stock-center（8081）、stock-manage（8082），健康检查通过后打开浏览器。若 `bin/` 下无二进制，将回退为 `go run`。

## 数据库备份

双库一并导出：

```bash
mysqldump -u root -p --databases stock_center stock_manage > backup-$(date +%Y%m%d).sql
```

还原：

```bash
mysql -u root -p < backup-YYYYMMDD.sql
```

## API（示例）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/stocks` | 分页列表 |
| GET | `/api/v1/stocks/:id` | 按 ID 查询 |
| GET | `/api/v1/stocks/by-sku?sku=&warehouse=` | 按 SKU + 仓库查询 |
| POST | `/api/v1/stocks` | 创建库存记录 |
| PUT | `/api/v1/stocks/:id/quantity` | 调整库存数量 |
