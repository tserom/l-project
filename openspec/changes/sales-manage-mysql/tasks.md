## 1. Scaffold sales-manage

- [ ] 1.1 新建 `apps/sales-manage`（go.mod、`cmd/server`、config、database、pkg/response），对齐 stock-manage 分层
- [ ] 1.2 配置 MySQL 库名 `sales_manage`、默认端口 `8083`、README 启动说明与 `.env.example`
- [ ] 1.3 根 `Makefile`（或脚本）增加 `dev-sales-manage` / 构建入口（若仓库已有同类目标则跟齐）

## 2. Schema and models

- [ ] 2.1 定义 `sales_order` / `sales_order_line` / `invoice_doc` / `invoice_doc_line` GORM model（字符串 UUID 对外 id、`invoice_doc_id` 可空）
- [ ] 2.2 AutoMigrate 与基本 repository CRUD（头行加载嵌套）

## 3. Query predicates (qp)

- [ ] 3.1 新增本服务 `internal/pkg/qp`（参考 stock-center）：解析 + 白名单 + 非法 400
- [ ] 3.2 销售单列表白名单：`customerName` like、`orderNo` eq/in/like、`orderDate` gte/lte、可选 `warehouseName` eq
- [ ] 3.3 开票列表白名单：`status` eq、`invoiceNo` like

## 4. Sales order HTTP API

- [ ] 4.1 实现 service：校验、合计、create/update/get/list/delete；更新保留 `invoiceDocId`；有开票挂接则禁止删除
- [ ] 4.2 注册路由 `/api/v1/sale/orders`（GET list/POST/GET:id/PUT:id/DELETE:id）；静态 `invoice-candidates` 稍后挂同一 group 且先于 `:id`
- [ ] 4.3 补充 Go 单测或 handler 级测试覆盖校验与删除拒绝

## 5. Invoice HTTP API

- [ ] 5.1 `GET /api/v1/sale/orders/invoice-candidates`：`needInvoice` 且无 `invoiceDocId` + qp 筛选
- [ ] 5.2 `POST /api/v1/sale/invoice-docs`：单事务写开票 + 回写销售行；冲突文案对齐前端
- [ ] 5.3 `POST /api/v1/sale/invoice-docs/:id/void`：单事务清挂接 + voided；拒绝重复作废
- [ ] 5.4 `GET` list/get invoice-docs（qp + 分页）
- [ ] 5.5 开票事务单测（成功 / 冲突回滚 / 作废）

## 6. sales-front HTTP cutover

- [ ] 6.1 增加 API base / Vite proxy 到 `8083`；薄 `httpClient`（若需要）
- [ ] 6.2 改写 `salesOrderApi`：全部走 `/api/v1/sale/orders`；列表参数编为 `qp-*` + 分页
- [ ] 6.3 改写 `invoiceDocApi`：list/get/candidates/create/void 走 HTTP；去掉 `db.transaction`
- [ ] 6.4 列表 / 候选页对接分页与 qp（停用主路径全量 `listAll`+内存滤）
- [ ] 6.5 移除 Dexie 运行时依赖与 `storage/*` 写路径；调整/删除 `fake-indexeddb` 相关测试，保留纯函数测试
- [ ] 6.6 更新 `sales-front` README 与仓内 `docs/domain-sales.md`、`docs/api-sales-storage.md` 真源说明

## 7. Docs and kb sync

- [ ] 7.1 更新 kb `sales-manage-api.md` / `local-first-to-go-mysql.md` / `personal-dev.md` / `stack.md` 为已落地事实（端口、启动命令）
- [ ] 7.2 确认文档写明：导入脚本第二期；与 stock-manage `/sales-orders` 无关

## 8. Manual verification

- [ ] 8.1 本地：建库 → 起 sales-manage → 起 sales-front → 销售单 CRUD、列表筛选、开票创建、作废、打印只读
- [ ] 8.2 确认浏览器不再出现 `sales-front` IndexedDB 作为写库（或库不再被打开）
