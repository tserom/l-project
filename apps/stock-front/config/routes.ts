/**
 * 路由配置（唯一真源）
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
      { path: 'materials', name: 'materials', label: '物料档案', component: './pages/materials' },
      { path: 'stocks', name: 'stocks', label: '库存查询', component: './pages/stocks' },
      { path: 'inbound', name: 'inbound', label: '入库单', component: './pages/inbound' },
      { path: 'outbound', name: 'outbound', label: '出库单', component: './pages/outbound' },
      { path: 'sales', name: 'sales', label: '销售单', component: './pages/sales' },
      { path: 'processing', name: 'processing', label: '加工单', component: './pages/processing' },
    ],
  },
]

export default routes
