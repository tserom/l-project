/**
 * 路由配置（唯一真源，与 Host 导航 apps[].routes 对齐）
 */
export interface RouteConfig {
  path?: string
  name?: string
  label?: string
  component?: string
  children?: RouteConfig[]
  redirect?: string
}

const routes: RouteConfig[] = [
  {
    path: '/',
    component: './layouts/BasicLayout',
    children: [
      { path: '', name: 'home', label: '首页', component: './pages/home' },
      { path: 'about', name: 'about', label: '关于', component: './pages/about' },
    ],
  },
]

export default routes
