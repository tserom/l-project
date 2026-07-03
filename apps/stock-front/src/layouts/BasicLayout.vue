<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="layout">
    <header class="layout__header">
      <nav class="layout__nav">
        <RouterLink to="/">首页</RouterLink>
        <RouterLink to="/about">关于</RouterLink>
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
  gap: 16px;
}

.layout__nav a {
  color: #374151;
  text-decoration: none;
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
