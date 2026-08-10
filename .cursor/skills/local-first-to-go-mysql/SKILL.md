---
name: local-first-to-go-mysql
description: >-
  个人项目默认交付路径：阶段一 IndexedDB(Dexie)+前端 *Api，阶段二 Go(Gin)+MySQL；
  开局检查表、qp-* 筛选契约、Flip 日单真源切换。
  Use when scaffolding a personal local-first app, migrating Dexie to Go,
  or reviewing sales-front-style architecture. Not for company BsSula/umi.
disable-model-invocation: false
---

# 本地优先 → Go + MySQL

## 何时读本 Skill

- 新个人工具：「先本地存、以后上后端」
- 从 Dexie / IndexedDB 迁到 Gin + MySQL
- 审查分层是否会导致后期难换
- 用户提到销售单式本地优先、Flip、单真源

**不要**用于公司 umi / BsSulaQueryTable 项目。

## 必读顺序

1. 团队 guideline：`guidelines/local-first-to-go-mysql.md`（MUST）
2. **完整检查表 / Flip**：知识库 `kb/domains/workflow/local-first-to-go-mysql.md`
3. Dexie 分层：`kb/domains/personal-react/indexeddb-dexie.md`
4. 列表筛选：`guidelines/api-query-predicate-params.md` + kb `api-query-predicate-params.md`

## 业务项目 LOCAL.md（建议）

复制本目录 `LOCAL.example.md` → 项目 `.cursor/skills/local-first-to-go-mysql/LOCAL.md`，填：

| 项 | 说明 |
|----|------|
| 当前阶段 | 一 / Flip 中 / 二 |
| 真源 | IndexedDB 库名 **或** Go 服务名/端口/库名 |
| Api 目录 | 如 `apps/xxx/src/services/` |
| 列表 qp 白名单文档路径 | |
| 跨实体写接口 | 阶段一事务函数名 → 阶段二 HTTP 路径 |
| 迁移 | 无数据直切 / 有导出导入（状态） |

## 实施步骤（摘要）

### A. 新项目开局（阶段一）

1. 建 `types/`、`services/*Api.ts`、`storage/*Repository.ts`、`storage/db.ts`
2. 页面只调 Api；写 README 真源与清站点风险
3. 为每个列表写 `field + operator` 白名单（实现可内存过滤）
4. 跨实体写集中在 Api + Dexie `transaction`
5. 用 kb §2 开局检查表打勾

### B. 迁后端（阶段二）

1. 建 Go app（Gin+GORM+MySQL）；表：头+行；DTO 对齐现有 Api JSON
2. List 落地 `qp-*`；非法 400
3. `*Api` 内改 fetch；环境开关若并存则**默认只开一种可写**
4. 跨表逻辑进单 HTTP 事务

### C. Flip 日

按 kb §6：空库直切 **或** 导出→导入→验证→删/归档 Dexie → 改文档真源。  
**禁止**双写、半切页面。

## 完成定义

- [ ] 页面无 Dexie/db import（阶段二）或仅 Api 下层使用（阶段一）
- [ ] 列表筛选语义与 `qp-*` 文档一致
- [ ] 唯一可写真源已文档化
- [ ] kb / README 范例状态已更新（若适用）
