## Why

经营不锈钢材料售卖与零件加工，需要记录入库单、出库单、销售单与加工损耗，现有 `stock-center` / `stock-manage` 仅为通用 SKU 整数库存，无法满足材质/炉号/双计量等业务。需在个人电脑本机 MySQL 8 部署，并为日后多业务与权限扩展保留双服务架构。

## What Changes

- **stock-center**：新增物料档案、炉号/批次、双计量库存余额、库存流水；新增 inbound/outbound API；**BREAKING** 废弃旧 `stock` 表与 `/api/v1/stocks` SKU 接口（迁移后移除）
- **stock-manage**：新增入库单、出库单、销售单、销售出库、加工单及确认编排；扩展 center 客户端；内嵌 `stock-front` 静态资源
- **stock-front**：新增物料、库存、四类单据页面；仅调用 manage API
- **部署**：新增启动器脚本、构建说明、MySQL 双库初始化文档
- 详细领域设计见 `docs/superpowers/specs/2026-07-03-stainless-steel-inventory-design.md`

## Capabilities

### New Capabilities

- `material-master`：物料档案与炉号/批次（stock-center）
- `stock-balance`：双计量库存余额、出入库、流水（stock-center）
- `business-documents`：入库/出库/销售/加工单据与确认编排（stock-manage）
- `local-deployment`：本机 MySQL 8、双进程启动器、可执行文件打包
- `inventory-ui`：Vue 前端页面与 manage BFF 接口

### Modified Capabilities

（无既有 openspec spec）

## Impact

- `apps/stock-center/**`：模型、仓储、服务、路由全面扩展
- `apps/stock-manage/**`：新单据域、center 客户端、静态资源 embed
- `apps/stock-front/**`：新页面与 API 层
- 新增 `scripts/launcher/**`、根目录或 `apps/` 下打包文档
- 依赖：Go `shopspring/decimal`；前端无新重型依赖
