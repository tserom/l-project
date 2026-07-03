<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const navItems = [
  { to: '/', label: '首页' },
  { to: '/materials', label: '物料档案' },
  { to: '/stocks', label: '库存查询' },
  { to: '/inbound', label: '入库单' },
  { to: '/outbound', label: '出库单' },
  { to: '/sales', label: '销售单' },
  { to: '/processing', label: '加工单' },
]
</script>

<template>
  <div class="layout">
    <header class="layout__header">
      <nav class="layout__nav">
        <RouterLink v-for="item in navItems" :key="item.to" :to="item.to">
          {{ item.label }}
        </RouterLink>
      </nav>
    </header>
    <main class="layout__main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.layout__header {
  padding: 12px 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #fff;
}

.layout__nav {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.layout__nav a {
  color: #374151;
  text-decoration: none;
  font-size: 0.9rem;
}

.layout__nav a.router-link-active {
  color: #42b883;
  font-weight: 600;
}

.layout__main {
  flex: 1;
  padding: 24px 20px;
}
</style>
