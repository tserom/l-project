# AGENTS.md

## Cursor Cloud specific instructions

Product: a stainless-steel inventory system (库存管理系统). Monorepo with three apps under `apps/`:

- `stock-center` (Go/Gin/GORM) — data-center service (materials, batches, stock, ledger). MySQL DB `stock_center`. Port `8081`.
- `stock-manage` (Go/Gin/GORM) — business BFF + serves the embedded frontend SPA. MySQL DB `stock_manage`, proxies stock data to `stock-center`. Port `8082`.
- `stock-front` (Vue 3 + Vite + TypeScript) — SPA source; built and embedded into `stock-manage` for production, or run standalone in dev on port `8104`.

### Startup (not in the update script — do this each session)

MySQL is required but must be started manually each session; the update script only refreshes dependencies.

- Start MySQL: `sudo service mysql start` (root is configured as `root`/`root` over TCP `127.0.0.1:3306`).
- Ensure DBs exist (idempotent): `mysql -h 127.0.0.1 -u root -proot < scripts/mysql/init-databases.sql`. Tables are auto-migrated by each Go service on startup (GORM `AutoMigrate`), so no manual migration step is needed.
- Run both backends: `./scripts/launcher/start.sh` (starts `stock-center`, waits for `:8081/health`, then `stock-manage`, waits for `:8082/health`; falls back to `go run` if `bin/` is absent). The script tries to open a browser via `xdg-open`; the resulting Chrome/DBus errors in its log are harmless noise.
- App UI: `http://localhost:8082`. Health: `GET /health` on `:8081` and `:8082`.

Config note: both Go services have built-in defaults (`internal/config/config.go`) that match `.env.example` exactly (root/root, ports 8081/8082, `STOCK_CENTER_BASE_URL=http://127.0.0.1:8081`), so `.env` files are optional for local dev. `.env` is gitignored.

### Frontend embedding (non-obvious)

`stock-manage` serves the SPA via `//go:embed internal/static/dist/*`. Only a placeholder `index.html` is committed (built `dist/*` assets are gitignored). To serve the real UI you must run `make build-front` (runs `pnpm build` in `stock-front` and copies `dist/` into the embed dir) before `make build-manage`. In pure frontend dev, run `pnpm --dir apps/stock-front dev` (Vite `:8104`) which proxies `/api` → `:8082`; the `stock-manage` backend must still be running.

### Build / lint / test (standard — see Makefiles)

- Build everything: `make build-all` (root `Makefile`). Per-app: `make build-front`, `make build-center`, `make build-manage`.
- Backends: `go -C apps/<svc> vet ./...`, `go -C apps/<svc> test ./...` (or `make -C apps/<svc> test`).
- Frontend type-check/lint: `pnpm --dir apps/stock-front type-check` (also part of `pnpm build`).

### Node version caveat

`apps/stock-front/package.json` pins `engines.node` to `>=20.19.5 <21`, but the VM's default Node is v22 and Vite 8 builds fine on it. `pnpm install`/`build` only print a non-fatal "Unsupported engine" warning — ignore it. There is no `.npmrc`, so `engine-strict` is off.
