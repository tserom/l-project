# 不锈钢库存系统（l-project）

本仓库为不锈钢库存管理本地部署方案：**stock-center**（数据中心）+ **stock-manage**（业务 API + 内嵌前端）+ **stock-front**（React 前端源码）。

```
浏览器 → stock-manage :8082（业务 + 静态页）
              │
              └── HTTP → stock-center :8081 → MySQL stock_center
              └── MySQL stock_manage（业务单据）
```

## 目录

| 路径 | 说明 |
|---|---|
| `apps/stock-center/` | 数据中心服务（物料、批次、库存、台账） |
| `apps/stock-manage/` | 业务服务（单据 + 代理 center + embed 前端） |
| `apps/stock-front/` | 前端源码（Vite + React） |
| `scripts/launcher/` | 一键启动 `start.sh` / `start.bat` |
| `scripts/pack/` | **Mac / Windows 打包脚本** |
| `scripts/mysql/` | 建库 SQL |
| `docs/windows-deployment.md` | Windows 部署与 API 清单 |

各子服务详见：`apps/stock-center/README.md`、`apps/stock-manage/README.md`、`apps/stock-front/README.md`。

## 开发构建

```bash
# 仓库根目录
make build-all          # 当前平台二进制 → bin/
make build-front        # 仅前端 + 复制到 manage embed 目录
make build-center       # 仅 stock-center
make build-manage       # 前端 + stock-manage
```

本地开发：先启 center，再启 manage；前端可 `cd apps/stock-front && pnpm dev`（代理 8082）。

## 打包分发（Mac / Windows）

在**有 Go、pnpm 的构建机**上执行（Windows 包可在 Mac/Linux 交叉编译）：

```bash
# Mac 运行时 zip（darwin arm64 或 amd64，按构建机架构）
make pack-mac
# 或：bash scripts/pack/pack-mac.sh

# Windows 运行时 zip（windows amd64，含 .exe）
make pack-windows
# 或：bash scripts/pack/pack-windows.sh
```

产物：

| 命令 | 输出 |
|---|---|
| `pack-mac` | `dist/stock-inventory-mac.zip` |
| `pack-windows` | `dist/stock-inventory-windows.zip` |

zip 内为**最小运行时包**（二进制 + `.env` 模板 + 启动器 + 建库 SQL），不含源码与 `node_modules`。

### 目标机首次使用

1. 解压 zip 到任意目录（保持内部目录结构）
2. 安装 MySQL 8，执行 `scripts/mysql/init-databases.sql`
3. 修改 `apps/stock-center/.env`、`apps/stock-manage/.env` 中的数据库密码
4. 启动：
   - **Mac**：`chmod +x scripts/launcher/start.sh && ./scripts/launcher/start.sh`
   - **Windows**：`scripts\launcher\start.bat`
5. 浏览器打开 **http://localhost:8082**

Windows 文件清单与完整 HTTP 接口见 [docs/windows-deployment.md](docs/windows-deployment.md)。

## 一键启动（源码目录）

已在本机构建过 `bin/` 时，可在仓库根目录直接：

```bash
# Mac / Linux
./scripts/launcher/start.sh

# Windows
scripts\launcher\start.bat
```

## 数据库备份

```bash
mysqldump -u root -p --databases stock_center stock_manage > backup-$(date +%Y%m%d).sql
```
