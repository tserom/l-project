import { createRouter, createWebHistory, type Router } from 'vue-router'
import BasicLayout from '@/layouts/BasicLayout.vue'

export function createAppRouter(): Router {
  return createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: '/',
        component: BasicLayout,
        children: [
          {
            path: '',
            name: 'home',
            component: () => import('@/pages/home/index.vue'),
          },
          {
            path: 'materials',
            name: 'materials',
            component: () => import('@/pages/materials/index.vue'),
          },
          {
            path: 'stocks',
            name: 'stocks',
            component: () => import('@/pages/stocks/index.vue'),
          },
          {
            path: 'inbound',
            name: 'inbound',
            component: () => import('@/pages/inbound/index.vue'),
          },
          {
            path: 'outbound',
            name: 'outbound',
            component: () => import('@/pages/outbound/index.vue'),
          },
          {
            path: 'sales',
            name: 'sales',
            component: () => import('@/pages/sales/index.vue'),
          },
          {
            path: 'processing',
            name: 'processing',
            component: () => import('@/pages/processing/index.vue'),
          },
        ],
      },
    ],
  })
}
