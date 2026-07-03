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
            path: 'about',
            name: 'about',
            component: () => import('@/pages/about/index.vue'),
          },
        ],
      },
    ],
  })
}
