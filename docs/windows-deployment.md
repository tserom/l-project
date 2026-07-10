# Windows 本机部署与文件清单

不锈钢库存系统（stock-center + stock-manage + 内嵌前端）在 Windows 上可通过 **预编译二进制 + 启动器** 一键运行，无需在目标机器安装 Go / Node。

---

## 一、运行前提

| 依赖 | 说明 |
|---|---|
| **MySQL 8** | 本机 `127.0.0.1:3306`，需先建库 |
| **curl** | `start.bat` 健康检查用；Windows 10 及以上通常已自带 |
| **Go**（可选） | 仅当 `bin/` 下无二进制、启动器回退 `go run` 时需要 |

---

## 二、在 Windows 上执行（推荐流程）

### 1. 初始化数据库（首次）

在 MySQL 客户端或命令行执行：

```bat
mysql -u root -p < scripts\mysql\init-databases.sql
```

会创建 `stock_center`、`stock_manage` 两个库；表结构由服务启动时 GORM AutoMigrate 自动建表。

### 2. 配置环境变量

```bat
copy apps\stock-center\.env.example apps\stock-center\.env
copy apps\stock-manage\.env.example apps\stock-manage\.env
```

按本机 MySQL 账号修改两个 `.env` 中的 `DB_PASSWORD` 等字段。`stock-manage` 的 `STOCK_CENTER_BASE_URL` 保持 `http://127.0.0.1:8081`。

### 3. 一键启动

在仓库根目录双击或命令行执行：

```bat
scripts\launcher\start.bat
```

启动器会：

1. 拉起 **stock-center**（端口 **8081**），等待 `/health` 就绪  
2. 拉起 **stock-manage**（端口 **8082**，内嵌前端），等待 `/health` 就绪  
3. 自动打开浏览器 **http://localhost:8082**

### 4. 验证

| 检查项 | 地址 |
|---|---|
| 数据中心健康检查 | http://localhost:8081/health |
| 业务服务 + 前端 | http://localhost:8082 |
| 业务服务健康检查 | http://localhost:8082/health |

---

## 三、需要复制出来的文件（运行时包）

将下列文件/目录按 **相同相对路径** 复制到 Windows 目标目录（下称 `<部署根>`）。启动器假定目录结构与仓库一致：`scripts\launcher\start.bat` 会 `cd` 到 `<部署根>`（即 `scripts` 的上两级）。

### 3.1 最小运行时包（给业务用户，无需源码）

```
<部署根>/
├── bin/
│   ├── stock-center.exe      # Windows 可执行文件（必须）
│   └── stock-manage.exe      # 已 embed 前端静态资源（必须）
├── apps/
│   ├── stock-center/
│   │   └── .env              # 由 .env.example 复制并修改（必须）
│   └── stock-manage/
│       └── .env              # 由 .env.example 复制并修改（必须）
└── scripts/
    ├── launcher/
    │   └── start.bat         # Windows 一键启动（必须）
    └── mysql/
        └── init-databases.sql # 首次建库用（首次部署必须）
```

**说明：**

- `stock-manage.exe` 已通过 `go:embed` 打包前端，**不需要**再复制 `apps/stock-front/dist` 或 `internal/static/dist`。
- `bin\` 下文件名须为 `stock-center.exe`、`stock-manage.exe`（`start.bat` 优先查找带 `.exe` 后缀）。
- **不要**复制 `node_modules`、`.git`、`apps/stock-front` 源码（运行时不需要）。

### 3.2 参考：从示例生成配置（开发机或打包机）

打包前可从仓库复制示例配置（若目标机尚未有 `.env`）：

| 源文件 | 复制为 |
|---|---|
| `apps/stock-center/.env.example` | `apps/stock-center/.env` |
| `apps/stock-manage/.env.example` | `apps/stock-manage/.env` |

### 3.3 开发/调试包（可选，含源码回退）

若希望在 Windows 上无二进制时用 `go run` 启动，需额外带上 Go 模块源码：

```
<部署根>/
├── apps/
│   ├── stock-center/         # 完整 Go 工程（go.mod、cmd、internal…）
│   └── stock-manage/         # 完整 Go 工程（含 internal/static/dist 若本地编译 manage）
├── bin/                      # 可为空；空则 start.bat 回退 go run
└── scripts/
    └── launcher/
        └── start.bat
```

本地在 Windows 编译 manage 时，需先构建前端并复制到 embed 目录（与 `Makefile` 一致）：

```bat
cd apps\stock-front
pnpm install
pnpm build
xcopy /E /I /Y dist ..\stock-manage\internal\static\dist

cd ..\stock-center
go build -o ..\..\bin\stock-center.exe .\cmd\server

cd ..\stock-manage
go build -o ..\..\bin\stock-manage.exe .\cmd\server
```

### 3.4 在 macOS/Linux 交叉编译 Windows 二进制

在开发机仓库根目录执行（需已 `make build-front` 或等价步骤，以便 manage embed 前端）：

```bash
# 构建前端并嵌入 manage（与 make build-manage 相同）
make build-front

cd apps/stock-center
GOOS=windows GOARCH=amd64 go build -o ../../bin/stock-center.exe ./cmd/server

cd ../stock-manage
GOOS=windows GOARCH=amd64 go build -o ../../bin/stock-manage.exe ./cmd/server
```

再将 **第三节 3.1** 所列目录打成 zip 拷贝到 Windows。

---

## 四、环境变量说明（`.env`）

### stock-center（`apps/stock-center/.env`）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_ENV` | development | 运行环境 |
| `SERVER_PORT` | 8081 | HTTP 端口 |
| `DB_HOST` | 127.0.0.1 | MySQL 主机 |
| `DB_PORT` | 3306 | MySQL 端口 |
| `DB_USER` | root | 数据库用户 |
| `DB_PASSWORD` | root | 数据库密码 |
| `DB_NAME` | stock_center | 库名 |
| `DB_CHARSET` | utf8mb4 | 字符集 |

### stock-manage（`apps/stock-manage/.env`）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_ENV` | development | 运行环境 |
| `SERVER_PORT` | 8082 | HTTP 端口（浏览器入口） |
| `DB_HOST` ~ `DB_CHARSET` | 同上 | 业务库 `stock_manage` |
| `STOCK_CENTER_BASE_URL` | http://127.0.0.1:8081 | 上游 stock-center 地址 |

---

## 五、HTTP 接口清单

对外业务入口为 **stock-manage `http://localhost:8082`**（前端与 REST API 同源）。  
stock-center `http://localhost:8081` 为数据中心，一般由 manage 内部调用；联调或排错时可直连。

### 5.1 健康检查

| 服务 | 方法 | 路径 |
|---|---|---|
| stock-center | GET | `/health` |
| stock-manage | GET | `/health` |

### 5.2 stock-center（8081）— `/api/v1`

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/materials` | 物料主数据分页列表 |
| GET | `/api/v1/materials/:id` | 物料详情 |
| POST | `/api/v1/materials` | 新建物料 |
| PUT | `/api/v1/materials/:id` | 更新物料 |
| GET | `/api/v1/batches` | 批次分页列表 |
| GET | `/api/v1/batches/:id` | 批次详情 |
| POST | `/api/v1/batches` | 新建批次 |
| PUT | `/api/v1/batches/:id` | 更新批次 |
| GET | `/api/v1/stocks` | 库存余额分页列表 |
| GET | `/api/v1/stocks/query` | 按物料+批次+仓库查单条库存 |
| POST | `/api/v1/stocks/inbound` | 库存入库（数据中心侧） |
| POST | `/api/v1/stocks/outbound` | 库存出库（数据中心侧） |
| GET | `/api/v1/ledger` | 库存台账分页列表 |

**列表通用分页参数：** `page`（默认 1）、`pageSize`（默认 20）

**`GET /api/v1/stocks/query` 查询参数：**

| 参数 | 必填 | 说明 |
|---|---|---|
| `materialId` | 是 | 物料 ID |
| `batchId` | 是 | 批次 ID |
| `warehouse` | 否 | 仓库编码 |

**`qp-*` 筛选（列表接口，多条件 AND）：**

| 接口 | 可用 `qp-<field>-<operator>` |
|---|---|
| `GET /materials` | `grade`（eq/like/in）、`form`（eq/like/in）、`materialType`（eq/like/in）、`status`（eq/like/in） |
| `GET /batches` | `materialId`（eq） |
| `GET /stocks` | `materialId`、`batchId`、`warehouse`（均 eq） |
| `GET /ledger` | `refNo`、`materialId`（均 eq） |

示例：`GET /api/v1/materials?qp-grade-eq=304&page=1&pageSize=20`

### 5.3 stock-manage（8082）— `/api/v1`

前端与第三方应调用 **manage** 层 API；物料/批次/库存/台账与 center 路径一致，由 manage **代理转发** 到 stock-center。

#### 代理（转发至 center）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/materials` | 物料列表（代理） |
| POST | `/api/v1/materials` | 新建物料（代理） |
| GET | `/api/v1/materials/:id` | 物料详情（代理） |
| PUT | `/api/v1/materials/:id` | 更新物料（代理） |
| GET | `/api/v1/batches` | 批次列表（代理） |
| POST | `/api/v1/batches` | 新建批次（代理） |
| GET | `/api/v1/batches/:id` | 批次详情（代理） |
| PUT | `/api/v1/batches/:id` | 更新批次（代理） |
| GET | `/api/v1/stocks` | 库存列表（代理） |
| GET | `/api/v1/stocks/query` | 库存查询（代理） |
| GET | `/api/v1/ledger` | 台账列表（代理） |

#### 业务单据（manage 本地库 + 确认时调 center）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/inbound-orders` | 入库单列表 |
| GET | `/api/v1/inbound-orders/:id` | 入库单详情 |
| POST | `/api/v1/inbound-orders` | 新建入库单 |
| PUT | `/api/v1/inbound-orders/:id` | 更新入库单 |
| DELETE | `/api/v1/inbound-orders/:id` | 删除入库单 |
| POST | `/api/v1/inbound-orders/:id/confirm` | 确认入库（写库存） |
| GET | `/api/v1/outbound-orders` | 出库单列表 |
| GET | `/api/v1/outbound-orders/:id` | 出库单详情 |
| POST | `/api/v1/outbound-orders` | 新建出库单 |
| PUT | `/api/v1/outbound-orders/:id` | 更新出库单 |
| DELETE | `/api/v1/outbound-orders/:id` | 删除出库单 |
| POST | `/api/v1/outbound-orders/:id/confirm` | 确认出库 |
| GET | `/api/v1/sales-orders` | 销售订单列表 |
| GET | `/api/v1/sales-orders/:id` | 销售订单详情 |
| POST | `/api/v1/sales-orders` | 新建销售订单 |
| PUT | `/api/v1/sales-orders/:id` | 更新销售订单 |
| DELETE | `/api/v1/sales-orders/:id` | 删除销售订单 |
| POST | `/api/v1/sales-orders/:id/confirm` | 确认销售订单 |
| GET | `/api/v1/sales-orders/:id/shipments` | 某销售单下的发货单列表 |
| POST | `/api/v1/sales-orders/:id/shipments` | 在某销售单下创建发货单 |
| GET | `/api/v1/sales-shipments` | 发货单列表 |
| GET | `/api/v1/sales-shipments/:id` | 发货单详情 |
| POST | `/api/v1/sales-shipments` | 新建发货单 |
| PUT | `/api/v1/sales-shipments/:id` | 更新发货单 |
| DELETE | `/api/v1/sales-shipments/:id` | 删除发货单 |
| POST | `/api/v1/sales-shipments/:id/confirm` | 确认发货（扣库存） |
| GET | `/api/v1/processing-orders` | 加工单列表 |
| GET | `/api/v1/processing-orders/:id` | 加工单详情 |
| POST | `/api/v1/processing-orders` | 新建加工单 |
| PUT | `/api/v1/processing-orders/:id` | 更新加工单 |
| DELETE | `/api/v1/processing-orders/:id` | 删除加工单 |
| POST | `/api/v1/processing-orders/:id/confirm` | 确认加工 |

**业务单据列表分页：** `page`、`pageSize`（同上，默认 1 / 20）。

#### 静态前端

| 说明 | 行为 |
|---|---|
| 非 `/api` 路径 | 由 manage 返回内嵌 SPA；未知路由回退 `index.html` |

---

## 六、常见问题

| 现象 | 处理 |
|---|---|
| 启动器报 MySQL / health 超时 | 确认 MySQL 已启动、`.env` 账号密码正确、8081/8082 端口未被占用 |
| 找不到 `stock-center.exe` | 确认 `bin\` 在部署根下，且为 Windows amd64 编译产物 |
| 浏览器 8082 无页面 | 须使用已 embed 前端的 `stock-manage.exe`（`make build-manage` 或第三节交叉编译流程） |
| 仅需 API、不需 UI | 仍建议启动 manage；也可只启 center（8081），但无 Web 界面 |

---

## 七、相关仓库文件索引

| 路径 | 用途 |
|---|---|
| `scripts/launcher/start.bat` | Windows 启动器 |
| `scripts/launcher/start.sh` | macOS / Linux 启动器 |
| `scripts/mysql/init-databases.sql` | 建库脚本 |
| `apps/stock-center/.env.example` | center 配置模板 |
| `apps/stock-manage/.env.example` | manage 配置模板 |
| `Makefile` | `build-all` / `build-manage` 构建命令 |
| `apps/stock-center/internal/router/router.go` | center 路由定义 |
| `apps/stock-manage/internal/router/router.go` | manage 路由定义 |
