# LOCAL.md — l-project 销售开票

| 项 | 本项目 |
|----|--------|
| 当前阶段 | **二**（IndexedDB 已退役） |
| 真源 | `apps/sales-manage` :8083 / MySQL `sales_manage` |
| 前端 Api 目录 | `apps/sales-front/src/services/` |
| 列表 qp 白名单 | kb `domains/personal-go/sales-manage-api.md`；后端 `apps/sales-manage` |
| 跨实体写 | 开票：原 Dexie `transaction(salesOrders, invoiceDocs)` → `POST/…` invoice-docs（见 sales-manage） |
| 数据迁移 | 导入脚本 **第二期**（无强制 JSON 导入） |
| 与其它服务 | 与 stock-\* **未打通**；勿用 stock-manage `/sales-orders` |
| 参考文档 | kb `domains/workflow/local-first-to-go-mysql.md` §8 |
