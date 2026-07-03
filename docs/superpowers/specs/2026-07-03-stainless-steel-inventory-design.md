# 不锈钢进销存系统设计

**日期**：2026-07-03  
**状态**：已评审（用户确认架构与单据方案）  
**范围**：第一版 — 本机 MySQL 8、双服务、单人使用，架构预留多人与上云

---

## 1. 背景与目标

### 1.1 业务场景

经营不锈钢材料售卖与零件加工，需同时管理：

- **原材料**：按重量（kg）计量
- **加工成品/零件**：按件数或长度计量
- **材质牌号、形态、规格、炉号/批次** 为库存核心维度

日常以 **销售单驱动出库**；入库、加工领料为独立流程，不与销售单强绑定。

加工环节：记录成品入库与投料出库；**不单独区分可利用边角料与纯废料**，以「投入重量 − 产出重量」体现损耗。

### 1.2 产品目标

| 目标 | 说明 |
|------|------|
| 单据管理 | 入库单、出库单、销售单、加工单 |
| 库存准确 | 双计量（重量 + 数量）、批次追踪、完整流水 |
| 本地部署 | 个人电脑运行，不租服务器；MySQL 8 本机安装 |
| 可打包分发 | 可执行文件 + 启动器，双击即用 |
| 可演进 | 先单人；预留组织、权限、上云 |

### 1.3 非目标（第一版不做）

- 完整 RBAC 与多租户
- 边角料/废料分类库存
- 财务对账、应收应付
- 云端同步、移动端

---

## 2. 架构

### 2.1 总体结构

```
浏览器
   │
   ▼
stock-front（Vue 3 静态资源，由 stock-manage 内嵌提供）
   │
   ▼
stock-manage :8082          stock-center :8081
业务层 · stock_manage DB  ──HTTP──▶  库存域 · stock_center DB
   │                                      │
   └────────────── MySQL 8 本机 ──────────┘
```

### 2.2 服务职责

| 服务 | 数据库 | 职责 | 禁止 |
|------|--------|------|------|
| **stock-center** | `stock_center` | 物料档案、批次/炉号、库存余额、库存流水；库存读写 API | 销售价、客户、单据、权限 |
| **stock-manage** | `stock_manage` | 入库/出库/销售/加工单及明细；业务审计；编排调用 center | 直接修改 center 库表 |
| **stock-front** | — | 仅调用 manage API | 直连 center |

### 2.3 演进路径

```
阶段 1（第一版）  本机 MySQL + 双进程 + 启动器；可无登录
阶段 2            局域网多浏览器；manage 增加用户/角色
阶段 3            云 MySQL + 服务器部署；manage ↔ center 内网 HTTP
```

### 2.4 与现有代码的关系

仓库已有 `stock-center`、`stock-manage`、`stock-front` 骨架（通用 SKU + 整数数量）。本设计在现有服务上**扩展域模型**，替换过于简化的 `Stock{SKU, Warehouse, Quantity}`，而非新建单体 `stock-app`。

---

## 3. 域模型（stock-center）

### 3.1 物料档案 `material`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| materialCode | string | 业务编码，唯一 |
| grade | string | 材质牌号：304、316L、201… |
| form | enum | 形态：板 / 管 / 棒 / 型材 / 零件 |
| spec | string | 规格描述（厚×宽×长、直径等） |
| primaryUnit | enum | 主单位：`kg` / `piece` / `meter` |
| secondaryUnit | enum? | 辅单位（可选） |
| materialType | enum | `raw` 原材料 / `finished` 成品 |
| status | enum | `enabled` / `disabled` |
| orgId | uint64? | 预留，第一版默认 0 |
| createdAt / updatedAt | time | |

### 3.2 批次/炉号 `material_batch`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| materialId | uint64 | 关联物料 |
| heatNo | string | 炉号/批次号 |
| remark | string? | |
| orgId | uint64? | 预留 |

同一 `materialId + heatNo` 在组织内唯一。

### 3.3 库存余额 `stock_balance`

按 **物料 + 批次 + 仓库** 一条记录，双计量：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| materialId | uint64 | |
| batchId | uint64 | |
| warehouse | string | 仓库编码，第一版可固定 `DEFAULT` |
| weightKg | decimal | 重量库存（kg） |
| quantity | decimal | 件数/长度库存 |
| orgId | uint64? | 预留 |

### 3.4 库存流水 `stock_ledger`（不可变）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | |
| materialId / batchId / warehouse | | |
| deltaWeightKg | decimal | 重量变动 ± |
| deltaQuantity | decimal | 数量变动 ± |
| refType | enum | `inbound` / `outbound` / `processing` / `sale` / `adjust` |
| refNo | string | 来源单号 |
| remark | string? | |
| createdAt | time | |

### 3.5 center 核心 API（第一版）

| 方法 | 路径 | 说明 |
|------|------|------|
| CRUD | `/api/v1/materials` | 物料档案 |
| CRUD | `/api/v1/batches` | 批次 |
| GET | `/api/v1/stocks` | 库存列表（分页、筛选） |
| GET | `/api/v1/stocks/query` | 按 materialId + batchId + warehouse 查询 |
| POST | `/api/v1/stocks/inbound` | 增库存（重量、数量可单独或同时传） |
| POST | `/api/v1/stocks/outbound` | 减库存（校验不足则 400） |
| GET | `/api/v1/ledger` | 流水查询 |

---

## 4. 域模型（stock-manage）

### 4.1 单据通用字段

所有单据头表共享：

| 字段 | 说明 |
|------|------|
| docNo | 单号，系统生成 |
| docDate | 业务日期 |
| status | `draft` 草稿 / `confirmed` 已确认 |
| operator | 经办人 |
| remark | 备注 |
| orgId | 预留 |
| createdBy | 预留 |
| createdAt / updatedAt | |

**规则**：仅 `confirmed` 状态触发库存变动；草稿可编辑、删除。

### 4.2 入库单 `inbound_order`

**场景**：采购来料、加工完工入库（非加工单确认时也可手工入库）。

明细行：

| 字段 | 说明 |
|------|------|
| materialId / batchId | 引用 center |
| warehouse | |
| weightKg | 入库重量（原材料必填） |
| quantity | 入库件数/长度（成品必填） |

确认后：调用 center `stocks/inbound`。

### 4.3 出库单 `outbound_order`

**场景**：加工领料、调拨、其它出库（非销售出库）。

明细行字段同入库。确认后：调用 center `stocks/outbound`。

### 4.4 销售单 `sales_order`

**场景**：对客户下单；**本身不扣库存**。

| 字段 | 说明 |
|------|------|
| customerName | 客户名（第一版简单文本） |
| 明细行 | materialId、batchId、销售数量/重量、单价（可选） |

### 4.5 销售出库 `sales_shipment`

由销售单生成，可分批。

| 字段 | 说明 |
|------|------|
| salesOrderId | 关联销售单 |
| 明细行 | 实际出库 materialId、batchId、weightKg、quantity |

确认后：调用 center `stocks/outbound`，`refType=sale`。

### 4.6 加工单 `processing_order`

一张单串联一次加工：

```
加工单头
├── 领料行（1..n）：原材料 materialId + batchId，出库 weightKg
└── 完工行（1..n）：成品 materialId + batchId（可新批次），入库 quantity（及可选 weightKg）
```

确认时 manage 编排：

1. 对每个领料行调用 center `outbound`（减重量）
2. 对每个完工行调用 center `inbound`（增数量/重量）
3. 计算并保存 `lossWeightKg = sum(领料 weightKg) − sum(完工 weightKg)`（完工无重量时按 0 计）

损耗仅记录在加工单上展示，**不单独入库**。

### 4.7 业务审计 `operation_log`

延续现有 `stock_operation_log` 思路，扩展为：单据类型、单号、动作、操作人、时间。

---

## 5. 关键业务流程

### 5.1 销售驱动出库

```
创建销售单（draft）→ 确认销售单 → 创建销售出库单 → 确认出库 → center 减库存
```

入库、加工不经过销售单。

### 5.2 采购/来料入库

```
创建入库单（draft）→ 录入明细 → 确认 → center 增库存
```

### 5.3 加工

```
创建加工单（draft）→ 填领料行 + 完工行 → 确认
  → center 出库（原料重量）
  → center 入库（成品件/米）
  → 记录 lossWeightKg
```

### 5.4 库存查询

前端 → manage → center 聚合查询；展示材质、形态、规格、炉号、重量、数量。

---

## 6. 本地部署与打包

### 6.1 环境依赖

- MySQL 8（用户本机安装）
- 数据库：`stock_center`、`stock_manage`（启动时自动迁移）

### 6.2 构建产物

| 产物 | 说明 |
|------|------|
| `stock-center.exe`（或 Mac 二进制） | `go build` |
| `stock-manage.exe` | `go build`，内嵌 `stock-front` 的 `dist/` |
| `启动.bat` / `启动.command` | 见 6.3 |
| `.env.example` | 数据库连接模板 |

### 6.3 启动器流程

1. 检查 MySQL 连通性（失败则提示安装/配置）
2. 启动 `stock-center`，轮询 `GET /health` 直至就绪
3. 启动 `stock-manage`，轮询 `GET /health`
4. 打开浏览器 `http://localhost:8082`

### 6.4 数据备份

文档说明使用 `mysqldump` 备份两库；建议定期复制备份文件。

### 6.5 上云迁移

- 将 MySQL 迁至云主机
- 修改两服务 `.env` 中的 `DB_HOST`
- 同一二进制部署；manage 配置 `STOCK_CENTER_URL` 指向 center 地址

---

## 7. 前端（stock-front）

### 7.1 第一版页面

| 页面 | 功能 |
|------|------|
| 首页 | 库存概览、快捷入口 |
| 物料档案 | 列表、新建/编辑（前端只调 manage，manage 转发 center） |
| 库存查询 | 按材质、形态、炉号筛选 |
| 入库单 | 列表、新建、确认 |
| 出库单 | 列表、新建、确认 |
| 销售单 | 列表、新建、确认、生成出库 |
| 加工单 | 列表、新建、确认、展示损耗 |

### 7.2 API 约定

- 浏览器只访问 manage（`:8082`）
- manage 对 center 的调用在服务端完成
- 列表筛选参数遵循项目 `qp-<field>-<operator>` 约定（后续实现时落地）

---

## 8. 预留扩展（表字段/API 预留，第一版不实现逻辑）

| 能力 | 预留方式 |
|------|----------|
| 多组织/业务线 | `orgId` 字段 |
| 用户与权限 | manage 用户表、`createdBy`、JWT |
| 服务间鉴权 | manage → center 携带 service token |
| 仓库/货位 | `warehouse` 字段扩展为货位表 |
| 供应商 | manage `supplier` 表 |
| 可利用边角料 | 新物料类型 + 加工单完工行类别 |

---

## 9. 风险与对策

| 风险 | 对策 |
|------|------|
| 双进程启动失败 | 启动器健康检查 + 明确错误提示 |
| 单据与库存不一致 | 仅 confirmed 调 center；center 出库校验余额 |
| 加工损耗无成品重量 | 允许完工行只填件数，损耗 = 投料重量 |
| 现有 Stock 模型冲突 | 新表与新 API 并行开发，迁移后废弃旧 `stock` 表 |

---

## 10. 成功标准（第一版）

- [ ] 本机 MySQL 8 + 启动器一键启动双服务与前端
- [ ] 物料档案含材质、形态、规格、炉号
- [ ] 入库单、出库单、销售单、加工单全流程可跑通
- [ ] 销售单确认后可销售出库并减库存
- [ ] 加工单确认后原料减重量、成品增件数，单上展示损耗
- [ ] 库存流水可追溯至来源单号

---

## 附录：需求确认记录

| 项 | 用户选择 |
|----|----------|
| 业务模式 | 卖料 + 加工都做 |
| 计量 | 原材料 kg，成品件/米 |
| 加工结果 | 成品入库；损耗 = 投入 − 产出，不分类边角料 |
| 物料维度 | 材质、形态、规格、炉号（必填） |
| 单据流程 | 销售单驱动出库；入库/加工独立 |
| 使用规模 | 先单人，架构可扩 |
| 数据库 | MySQL 8 本机（不用 SQLite） |
| 架构 | stock-center + stock-manage + 启动器 |
| 物料档案归属 | stock-center |
