# vue-project

Vue 3 + Vite 微前端子应用，已接入 [wujie](https://wujie-micro.github.io/doc/)（无界）。

## 本地开发

```bash
nvm use 20
cd apps/vue-project
pnpm install
pnpm dev
```

独立访问：<http://localhost:8104/>

## 接入 Host

| 项 | 值 |
|---|---|
| 导航 app key | `vue` |
| subAppBusName | `vue-project` |
| 父→子 bus 事件 | `vue-project-route` |
| 子→父 bus 事件 | `sub-route-change` |
| 同源 entry | `/micro/vue/` |
| dev 端口 | **8104** |

多端口联调时同时启动 `apps/host`（8100）与本应用；Host `.env.development` 中 `VITE_VUE_PROJECT_URL=http://localhost:8104/`。

## wujie 适配要点

- `vite.config.ts`：`base: './'`（子应用静态资源相对路径）
- `src/main.ts`：检测 `__POWERED_BY_WUJIE__`，注册 `__WUJIE_MOUNT` / `__WUJIE_UNMOUNT`
- `src/components/WujieRouteBridge.vue`：父子路由同步（initialPath、bus 事件）
- `src/utils/wujie.ts`：`SUB_APP_NAME` 须与导航 `subAppBusName` 一致

## 构建

```bash
pnpm build
pnpm preview
```

Docker 见根目录 `infra/docker/docker-compose.yml` 中的 `vue-project` 服务。
