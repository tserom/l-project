import { createApp, type App as VueApp } from 'vue'
import App from './App.vue'
import { createAppRouter } from './router'
import { IS_WUJIE } from './utils/wujie'
import './styles/global.css'

let app: VueApp | null = null

const renderApp = () => {
  const el = document.getElementById('app')
  if (!el) {
    throw new Error('[main] 找不到挂载节点 #app')
  }
  app = createApp(App)
  app.use(createAppRouter())
  app.mount(el)
}

const unmountApp = () => {
  app?.unmount()
  app = null
}

if (IS_WUJIE) {
  window.__WUJIE_MOUNT = renderApp
  window.__WUJIE_UNMOUNT = unmountApp
} else {
  renderApp()
}
