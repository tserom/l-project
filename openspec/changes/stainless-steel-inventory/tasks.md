## 1. 基础设施与依赖

- [x] 1.1 两服务 `go.mod` 添加 `github.com/shopspring/decimal`
- [x] 1.2 更新 `apps/stock-center/.env.example` 与 `apps/stock-manage/.env.example` 文档说明双库
- [x] 1.3 新增 `scripts/mysql/init-databases.sql` 创建 `stock_center`、`stock_manage`

## 2. stock-center 域模型

- [x] 2.1 新增 `internal/model/material.go`、`material_batch.go`、`stock_balance.go`、`stock_ledger.go`
- [x] 2.2 更新 `internal/database/mysql.go` AutoMigrate 新表，移除旧 `Stock` 迁移
- [x] 2.3 删除 `internal/model/stock.go` 及旧 stock repository/service/handler

## 3. stock-center 物料与批次 API

- [x] 3.1 实现 `material_repository`、`material_service`、`material_handler`
- [x] 3.2 实现 `batch_repository`、`batch_service`、`batch_handler`
- [x] 3.3 注册路由 `/api/v1/materials`、`/api/v1/batches`，列表支持 qp 筛选
- [x] 3.4 为 material/batch 编写 handler 级测试或 `go test ./internal/service/...`

## 4. stock-center 库存与流水 API

- [x] 4.1 实现 `stock_balance_repository`、`stock_ledger_repository`
- [x] 4.2 实现 `stock_balance_service`：`Inbound`、`Outbound`、`Query`、`List`
- [x] 4.3 实现 `stock_balance_handler` 与 `/api/v1/stocks`、`/api/v1/stocks/inbound`、`/api/v1/stocks/outbound`、`/api/v1/ledger`
- [x] 4.4 出库余额不足返回 400；流水写入与余额更新同一事务

## 5. stock-manage center 客户端

- [x] 5.1 扩展 `internal/client/stockcenter/client.go`：Material、Batch、StockBalance、Inbound、Outbound、Ledger 类型与方法
- [x] 5.2 删除旧 SKU `Stock` 客户端方法与 `inventory_service` 旧实现

## 6. stock-manage 单据模型与仓储

- [x] 6.1 新增模型：`inbound_order`、`outbound_order`、`sales_order`、`sales_shipment`、`processing_order` 及行表
- [x] 6.2 新增 `doc_sequence` 与 `doc_no` 生成器 `internal/pkg/docno`
- [x] 6.3 各单据 `repository` CRUD 与按状态查询

## 7. stock-manage 单据服务

- [x] 7.1 `inbound_service`：草稿 CRUD、Confirm 编排 center inbound
- [x] 7.2 `outbound_service`：草稿 CRUD、Confirm 编排 center outbound
- [x] 7.3 `sales_service`：销售单确认不扣库存；`shipment_service` 从销售单生成并确认出库
- [x] 7.4 `processing_service`：Confirm 先 outbound 领料再 inbound 完工，计算 `lossWeightKg`
- [x] 7.5 扩展 `operation_log` 记录单据动作

## 8. stock-manage HTTP 与 BFF

- [x] 8.1 新增 handler/router：物料与批次代理（转发 center）
- [x] 8.2 新增各单据 handler/router，统一 `pkg/response` 信封
- [x] 8.3 列表接口落地 qp 谓词（grade、form、status、docNo、docDate 等）

## 9. stock-front 前端

- [x] 9.1 移除 wujie 依赖与相关组件（第一版独立应用）
- [x] 9.2 新增 `src/api/manage.ts` 与类型定义
- [x] 9.3 页面：物料档案、库存查询、入库单、出库单、销售单（含出库）、加工单
- [x] 9.4 更新 `config/routes.ts`、`BasicLayout` 导航
- [x] 9.5 `vite.config.ts` 开发代理到 `http://localhost:8082`

## 10. 内嵌静态资源与打包

- [x] 10.1 `apps/stock-manage` 添加 `//go:embed` 与 `internal/static` 服务
- [x] 10.2 根目录或 `scripts/Makefile` 添加 `build-all`：前端 build → copy dist → go build 两二进制
- [x] 10.3 `scripts/launcher/start.sh`、`start.bat` 健康检查与打开浏览器
- [x] 10.4 更新 `apps/stock-center/README.md`、`apps/stock-manage/README.md` 部署与备份说明

## 11. 验收

- [ ] 11.1 手工走通：来料入库 → 加工 → 销售出库，库存与流水一致
- [x] 11.2 加工单展示 lossWeightKg；出库超库存返回错误（单元测试已覆盖；见 e2e 文档）
- [ ] 11.3 启动器在本机 MySQL 8 下一键启动通过
