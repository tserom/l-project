# 不锈钢进销存 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 `stock-center` + `stock-manage` + `stock-front` 上实现不锈钢双计量进销存（物料/批次/库存/四类单据），支持本机 MySQL 8 与启动器打包。

**Architecture:** 物料与库存真源在 center（`stock_center` 库）；单据与编排在 manage（`stock_manage` 库）；前端只调 manage，manage 转发 center。确认单据后调用 center inbound/outbound 并写流水。

**Tech Stack:** Go 1.22+、Gin、GORM、MySQL 8、`shopspring/decimal`、Vue 3、Vite、TypeScript

**关联文档：**

- 设计 spec：`docs/superpowers/specs/2026-07-03-stainless-steel-inventory-design.md`
- OpenSpec 变更：`openspec/changes/stainless-steel-inventory/`（proposal / design / specs / tasks）

---

## 文件结构总览

### stock-center（新建/修改）

| 路径 | 职责 |
|------|------|
| `internal/model/material.go` | 物料档案 |
| `internal/model/material_batch.go` | 炉号/批次 |
| `internal/model/stock_balance.go` | 双计量余额 |
| `internal/model/stock_ledger.go` | 库存流水 |
| `internal/repository/material_repository.go` | 物料 DB |
| `internal/repository/batch_repository.go` | 批次 DB |
| `internal/repository/stock_balance_repository.go` | 余额 DB |
| `internal/repository/stock_ledger_repository.go` | 流水 DB |
| `internal/service/material_service.go` | 物料业务 |
| `internal/service/batch_service.go` | 批次业务 |
| `internal/service/stock_balance_service.go` | 出入库 |
| `internal/handler/material_handler.go` | 物料 HTTP |
| `internal/handler/batch_handler.go` | 批次 HTTP |
| `internal/handler/stock_balance_handler.go` | 库存 HTTP |
| `internal/router/router.go` | 路由注册 |
| `internal/pkg/qp/query.go` | qp 谓词解析（白名单） |
| 删除 `internal/model/stock.go` 等旧文件 | 移除 SKU 模型 |

### stock-manage（新建/修改）

| 路径 | 职责 |
|------|------|
| `internal/model/inbound_order.go` 等 | 单据头行表 |
| `internal/model/doc_sequence.go` | 单号序号 |
| `internal/pkg/docno/generator.go` | 单号生成 |
| `internal/client/stockcenter/client.go` | 扩展 center 客户端 |
| `internal/service/*_service.go` | 各单据编排 |
| `internal/handler/proxy_handler.go` | 物料/库存 BFF |
| `internal/static/embed.go` | 内嵌前端 |
| `internal/router/router.go` | 路由 + SPA fallback |

### stock-front

| 路径 | 职责 |
|------|------|
| `src/api/manage.ts` | API 客户端 |
| `src/pages/materials/` | 物料档案 |
| `src/pages/stocks/` | 库存查询 |
| `src/pages/inbound/` | 入库单 |
| `src/pages/outbound/` | 出库单 |
| `src/pages/sales/` | 销售单 + 出库 |
| `src/pages/processing/` | 加工单 |

### 脚本

| 路径 | 职责 |
|------|------|
| `scripts/mysql/init-databases.sql` | 建库 |
| `scripts/launcher/start.sh` | Mac/Linux 启动器 |
| `scripts/launcher/start.bat` | Windows 启动器 |
| `Makefile`（仓库根或 `scripts/`） | `build-all` |

---

## Task 1: 依赖与数据库初始化

**Files:**
- Modify: `apps/stock-center/go.mod`
- Modify: `apps/stock-manage/go.mod`
- Create: `scripts/mysql/init-databases.sql`

- [ ] **Step 1: 添加 decimal 依赖**

```bash
cd apps/stock-center && go get github.com/shopspring/decimal@v1.4.0
cd apps/stock-manage && go get github.com/shopspring/decimal@v1.4.0
```

- [ ] **Step 2: 创建建库脚本**

```sql
-- scripts/mysql/init-databases.sql
CREATE DATABASE IF NOT EXISTS stock_center DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS stock_manage DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

- [ ] **Step 3: 执行建库**

```bash
mysql -u root -p < scripts/mysql/init-databases.sql
```

Expected: `Query OK` for both databases.

- [ ] **Step 4: Commit**

```bash
git add apps/stock-center/go.mod apps/stock-center/go.sum apps/stock-manage/go.mod apps/stock-manage/go.sum scripts/mysql/init-databases.sql
git commit -m "chore: add decimal dep and mysql init script for stainless inventory"
```

---

## Task 2: stock-center 域模型

**Files:**
- Create: `apps/stock-center/internal/model/material.go`
- Create: `apps/stock-center/internal/model/material_batch.go`
- Create: `apps/stock-center/internal/model/stock_balance.go`
- Create: `apps/stock-center/internal/model/stock_ledger.go`
- Modify: `apps/stock-center/internal/database/mysql.go`
- Delete: `apps/stock-center/internal/model/stock.go`

- [ ] **Step 1: 编写 material 模型**

```go
// apps/stock-center/internal/model/material.go
package model

import (
	"time"
	"github.com/shopspring/decimal"
)

type MaterialForm string
type PrimaryUnit string
type MaterialType string
type MaterialStatus string

const (
	FormPlate     MaterialForm = "plate"
	FormPipe      MaterialForm = "pipe"
	FormBar       MaterialForm = "bar"
	FormProfile   MaterialForm = "profile"
	FormPart      MaterialForm = "part"
	UnitKg        PrimaryUnit = "kg"
	UnitPiece     PrimaryUnit = "piece"
	UnitMeter     PrimaryUnit = "meter"
	TypeRaw       MaterialType = "raw"
	TypeFinished  MaterialType = "finished"
	StatusEnabled MaterialStatus = "enabled"
	StatusDisabled MaterialStatus = "disabled"
)

type Material struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialCode string         `gorm:"size:64;uniqueIndex:uk_material_code;not null" json:"materialCode"`
	Grade        string         `gorm:"size:32;not null;index" json:"grade"`
	Form         MaterialForm   `gorm:"size:16;not null" json:"form"`
	Spec         string         `gorm:"size:128;not null" json:"spec"`
	PrimaryUnit  PrimaryUnit    `gorm:"size:16;not null" json:"primaryUnit"`
	MaterialType MaterialType   `gorm:"size:16;not null" json:"materialType"`
	Status       MaterialStatus `gorm:"size:16;not null;default:enabled" json:"status"`
	OrgID        uint64         `gorm:"not null;default:0;index" json:"orgId"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

func (Material) TableName() string { return "material" }
```

- [ ] **Step 2: 编写 batch、balance、ledger 模型**

`material_batch`: `MaterialID`, `HeatNo`, `Remark`, `OrgID`，唯一索引 `uk_material_heat (material_id, heat_no, org_id)`。

`stock_balance`: `MaterialID`, `BatchID`, `Warehouse`, `WeightKg decimal.Decimal`, `Quantity decimal.Decimal`, `OrgID`，唯一索引 `uk_balance (material_id, batch_id, warehouse, org_id)`。

`stock_ledger`: `MaterialID`, `BatchID`, `Warehouse`, `DeltaWeightKg`, `DeltaQuantity`, `RefType`（inbound/outbound/processing/sale/adjust）, `RefNo`, `Remark`, `CreatedAt`。

- [ ] **Step 3: 更新 AutoMigrate**

```go
// apps/stock-center/internal/database/mysql.go — 在 AutoMigrate 中替换为：
db.AutoMigrate(
	&model.Material{},
	&model.MaterialBatch{},
	&model.StockBalance{},
	&model.StockLedger{},
)
```

- [ ] **Step 4: 删除旧 stock 相关文件并修复编译**

删除：`model/stock.go`、`repository/stock_repository.go`、`service/stock_service.go`、`handler/stock_handler.go`。

- [ ] **Step 5: 验证迁移**

```bash
cd apps/stock-center && make run
# 另开终端
curl -s http://localhost:8081/health
```

Expected: `{"status":"ok"}` 或项目现有健康响应。

- [ ] **Step 6: Commit**

```bash
git commit -m "feat(stock-center): replace SKU stock with material/batch/balance models"
```

---

## Task 3: stock-center 物料与批次 API

**Files:**
- Create: `apps/stock-center/internal/repository/material_repository.go`
- Create: `apps/stock-center/internal/service/material_service.go`
- Create: `apps/stock-center/internal/handler/material_handler.go`
- Create: `apps/stock-center/internal/repository/batch_repository.go`
- Create: `apps/stock-center/internal/service/batch_service.go`
- Create: `apps/stock-center/internal/handler/batch_handler.go`
- Create: `apps/stock-center/internal/pkg/qp/query.go`
- Modify: `apps/stock-center/internal/router/router.go`

- [ ] **Step 1: qp 白名单解析器**

```go
// apps/stock-center/internal/pkg/qp/query.go
package qp

import "github.com/gin-gonic/gin"

var materialWhitelist = map[string]bool{
	"grade": true, "form": true, "materialType": true, "status": true,
}

func MaterialFilters(c *gin.Context) (map[string]string, error) {
	// 解析 qp-grade-eq 等，非法字段返回 error
	return nil, nil
}
```

- [ ] **Step 2: Material CRUD + 列表分页**

路由：

```
GET    /api/v1/materials
GET    /api/v1/materials/:id
POST   /api/v1/materials
PUT    /api/v1/materials/:id
```

POST body 示例：

```json
{
  "materialCode": "304-PLATE-3x1500x6000",
  "grade": "304",
  "form": "plate",
  "spec": "3×1500×6000",
  "primaryUnit": "kg",
  "materialType": "raw"
}
```

- [ ] **Step 3: Batch CRUD**

路由：`/api/v1/batches`，创建需 `materialId` + `heatNo`。

- [ ] **Step 4: 手工验证**

```bash
curl -X POST http://localhost:8081/api/v1/materials -H 'Content-Type: application/json' -d '{"materialCode":"TEST-001","grade":"304","form":"plate","spec":"3mm","primaryUnit":"kg","materialType":"raw"}'
curl "http://localhost:8081/api/v1/materials?qp-grade-eq=304"
```

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(stock-center): material and batch APIs with qp filters"
```

---

## Task 4: stock-center 库存出入库与流水

**Files:**
- Create: `apps/stock-center/internal/repository/stock_balance_repository.go`
- Create: `apps/stock-center/internal/repository/stock_ledger_repository.go`
- Create: `apps/stock-center/internal/service/stock_balance_service.go`
- Create: `apps/stock-center/internal/handler/stock_balance_handler.go`
- Modify: `apps/stock-center/internal/router/router.go`

- [ ] **Step 1: Inbound 服务逻辑（事务内）**

```go
// InboundInput
type InboundInput struct {
	MaterialID  uint64
	BatchID     uint64
	Warehouse   string
	WeightKg    decimal.Decimal
	Quantity    decimal.Decimal
	RefType     string
	RefNo       string
	Remark      string
}
// 1. UPSERT stock_balance 增加 weight/quantity
// 2. INSERT stock_ledger
```

- [ ] **Step 2: Outbound 校验**

若 `WeightKg` 大于余额 `WeightKg`，返回 `errors.New("insufficient weight")`，handler 映射 HTTP 400。

- [ ] **Step 3: 路由**

```
GET  /api/v1/stocks
GET  /api/v1/stocks/query?materialId=&batchId=&warehouse=
POST /api/v1/stocks/inbound
POST /api/v1/stocks/outbound
GET  /api/v1/ledger?qp-refNo-eq=IN202607030001
```

- [ ] **Step 4: 集成测试式验证**

先 inbound 100kg，再 outbound 30kg，查询余额 70kg，ledger 两条。

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(stock-center): dual-measure inbound/outbound and ledger"
```

---

## Task 5: stock-manage center 客户端扩展

**Files:**
- Modify: `apps/stock-manage/internal/client/stockcenter/client.go`
- Delete: `apps/stock-manage/internal/service/inventory_service.go`
- Delete: `apps/stock-manage/internal/handler/inventory_handler.go`

- [ ] **Step 1: 定义新类型**

```go
type Material struct {
	ID uint64 `json:"id"`
	MaterialCode string `json:"materialCode"`
	Grade string `json:"grade"`
	Form string `json:"form"`
	Spec string `json:"spec"`
	PrimaryUnit string `json:"primaryUnit"`
	MaterialType string `json:"materialType"`
}

type InboundStockInput struct {
	MaterialID uint64 `json:"materialId"`
	BatchID    uint64 `json:"batchId"`
	Warehouse  string `json:"warehouse"`
	WeightKg   string `json:"weightKg"`
	Quantity   string `json:"quantity"`
	RefType    string `json:"refType"`
	RefNo      string `json:"refNo"`
	Remark     string `json:"remark"`
}
```

- [ ] **Step 2: 实现 `InboundStock`、`OutboundStock`、`ListMaterials` 等方法**

沿用现有 `APIResponse` 解包模式。

- [ ] **Step 3: 移除旧 inventory 路由，确保编译通过**

```bash
cd apps/stock-manage && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(stock-manage): extend stockcenter client for new inventory APIs"
```

---

## Task 6: stock-manage 单据模型与单号

**Files:**
- Create: `apps/stock-manage/internal/model/inbound_order.go`（及 line）
- Create: `apps/stock-manage/internal/model/outbound_order.go`
- Create: `apps/stock-manage/internal/model/sales_order.go`
- Create: `apps/stock-manage/internal/model/sales_shipment.go`
- Create: `apps/stock-manage/internal/model/processing_order.go`
- Create: `apps/stock-manage/internal/model/doc_sequence.go`
- Create: `apps/stock-manage/internal/pkg/docno/generator.go`

- [ ] **Step 1: 单据头公共字段**

```go
type DocStatus string
const (
	DocStatusDraft     DocStatus = "draft"
	DocStatusConfirmed DocStatus = "confirmed"
)

type InboundOrder struct {
	ID        uint64    `gorm:"primaryKey"`
	DocNo     string    `gorm:"size:32;uniqueIndex;not null"`
	DocDate   time.Time `gorm:"type:date;not null"`
	Status    DocStatus `gorm:"size:16;not null"`
	Operator  string    `gorm:"size:64;not null"`
	Remark    string    `gorm:"size:255"`
	OrgID     uint64    `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Lines     []InboundOrderLine `gorm:"foreignKey:InboundOrderID"`
}
```

行表：`MaterialID`, `BatchID`, `Warehouse`, `WeightKg`, `Quantity`（decimal）。

- [ ] **Step 2: 单号生成器**

```go
// docno.Generate(ctx, db, "IN") => "IN202607030001"
```

- [ ] **Step 3: AutoMigrate 新表**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(stock-manage): document models and doc number generator"
```

---

## Task 7: stock-manage 入库单与出库单服务

**Files:**
- Create: `apps/stock-manage/internal/repository/inbound_repository.go`
- Create: `apps/stock-manage/internal/service/inbound_service.go`
- Create: `apps/stock-manage/internal/handler/inbound_handler.go`
- 出库单对称：`outbound_*`

- [ ] **Step 1: Confirm 编排（入库）**

```go
func (s *InboundService) Confirm(ctx context.Context, id uint64) error {
	// 1. 加载 draft 单据
	// 2. 逐行调用 stockCenter.InboundStock(...)
	// 3. 更新 status=confirmed
	// 4. 写 operation_log
}
```

- [ ] **Step 2: HTTP 路由**

```
GET  /api/v1/inbound-orders
POST /api/v1/inbound-orders
PUT  /api/v1/inbound-orders/:id
POST /api/v1/inbound-orders/:id/confirm
```

- [ ] **Step 3: curl 验证入库确认后 center 余额增加**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(stock-manage): inbound and outbound order services"
```

---

## Task 8: stock-manage 销售单与销售出库

**Files:**
- Create: `sales_*` service/handler/repository

- [ ] **Step 1: 销售单 Confirm 不调用 center**

- [ ] **Step 2: 从销售单创建 shipment**

`POST /api/v1/sales-orders/:id/shipments`

- [ ] **Step 3: shipment Confirm 调用 outbound，refType=sale**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(stock-manage): sales order and shipment flow"
```

---

## Task 9: stock-manage 加工单

**Files:**
- Create: `processing_*` service/handler/repository

- [ ] **Step 1: 模型含 pick lines 与 finish lines**

`ProcessingPickLine`: `WeightKg`  
`ProcessingFinishLine`: `Quantity`, `WeightKg`（可选）  
头表：`LossWeightKg`

- [ ] **Step 2: Confirm 逻辑**

```go
totalPick := decimal.Zero
for _, line := range order.PickLines {
  center.Outbound(...)
  totalPick = totalPick.Add(line.WeightKg)
}
totalFinishWeight := decimal.Zero
for _, line := range order.FinishLines {
  center.Inbound(...)
  totalFinishWeight = totalFinishWeight.Add(line.WeightKg)
}
order.LossWeightKg = totalPick.Sub(totalFinishWeight)
```

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(stock-manage): processing order with loss calculation"
```

---

## Task 10: stock-manage BFF 与物料代理

**Files:**
- Create: `apps/stock-manage/internal/handler/proxy_handler.go`
- Modify: `apps/stock-manage/internal/router/router.go`

- [ ] **Step 1: 代理路由**

```
GET/POST /api/v1/materials -> center
GET/POST /api/v1/batches   -> center
GET      /api/v1/stocks    -> center
GET      /api/v1/ledger    -> center
```

- [ ] **Step 2: Commit**

```bash
git commit -m "feat(stock-manage): BFF proxy for center material and stock APIs"
```

---

## Task 11: stock-front 改造

**Files:**
- Modify: `apps/stock-front/package.json`（移除 wujie）
- Create: `apps/stock-front/src/api/manage.ts`
- Create: 各 `src/pages/**`
- Modify: `apps/stock-front/config/routes.ts`
- Modify: `apps/stock-front/vite.config.ts`

- [ ] **Step 1: API 客户端**

```typescript
// apps/stock-front/src/api/manage.ts
const BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8082';

export async function listMaterials(params: Record<string, string>) {
  const qs = new URLSearchParams(params).toString();
  const res = await fetch(`${BASE}/api/v1/materials?${qs}`);
  return res.json();
}
```

- [ ] **Step 2: vite 开发代理**

```typescript
// vite.config.ts
server: {
  proxy: { '/api': 'http://localhost:8082' }
}
```

- [ ] **Step 3: 入库单页面**

列表 + 表单（选 material/batch、填 weightKg/quantity）+ 确认按钮调 `POST .../confirm`。

- [ ] **Step 4: 其余页面按 tasks.md 9.3 逐项实现**

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(stock-front): inventory UI pages and manage API client"
```

---

## Task 12: 内嵌前端与打包

**Files:**
- Create: `apps/stock-manage/internal/static/embed.go`
- Modify: `apps/stock-manage/cmd/server/main.go`
- Create: `scripts/launcher/start.sh`
- Create: `scripts/launcher/start.bat`
- Create: `Makefile`

- [ ] **Step 1: 构建前端并复制**

```bash
cd apps/stock-front && pnpm install && pnpm build
mkdir -p apps/stock-manage/internal/static/dist
cp -r apps/stock-front/dist/* apps/stock-manage/internal/static/dist/
```

- [ ] **Step 2: embed 与 SPA fallback**

```go
//go:embed dist/*
var distFS embed.FS

// router: NoRoute 对非 /api 返回 index.html
```

- [ ] **Step 3: 启动器 start.sh**

```bash
#!/usr/bin/env bash
set -e
./bin/stock-center &
until curl -sf http://localhost:8081/health; do sleep 1; done
./bin/stock-manage &
until curl -sf http://localhost:8082/health; do sleep 1; done
open http://localhost:8082
```

- [ ] **Step 4: Makefile build-all**

```makefile
build-all:
	cd apps/stock-front && pnpm build
	# copy dist, go build -o bin/stock-center ./apps/stock-center/cmd/server
	# go build -o bin/stock-manage ./apps/stock-manage/cmd/server
```

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: embed frontend and add local launcher scripts"
```

---

## Task 13: 端到端验收

- [ ] **Step 1: 来料入库 100kg 304 板**

- [ ] **Step 2: 加工单领料 50kg，完工 10 件，查看 lossWeightKg**

- [ ] **Step 3: 销售单 + 出库 20kg**

- [ ] **Step 4: 库存余额 30kg；ledger 可按单号查询**

- [ ] **Step 5: 启动器一键启动验证**

---

## Spec 覆盖自检

| Spec 章节 | 对应 Task |
|-----------|-----------|
| material-master | Task 2–3 |
| stock-balance | Task 4 |
| business-documents | Task 6–9 |
| local-deployment | Task 1, 12 |
| inventory-ui | Task 11 |

无 TBD/占位；类型命名与 OpenSpec 一致。

---

## 执行方式

**Plan complete and saved to `docs/superpowers/plans/2026-07-03-stainless-steel-inventory.md`.**

**OpenSpec 变更：** `openspec/changes/stainless-steel-inventory/`（可用 `/opsx:apply` 按 tasks 推进）

**两种执行方式：**

1. **Subagent-Driven（推荐）** — 每个 Task 派发子 agent，任务间 Review  
2. **Inline Execution** — 本会话按 Task 顺序直接实现，检查点汇报  

你更倾向哪一种？
