<script setup lang="ts">
import { IS_WUJIE } from '@/utils/wujie'
</script>

<template>
  <section class="page">
    <h1>Vue 子应用 · 首页</h1>
    <p>这是一个 Vue 3 微前端子应用，已接入 wujie。</p>
    <p class="page__meta">
      运行模式：<strong>{{ IS_WUJIE ? '无界子应用' : '独立运行' }}</strong>
    </p>
  </section>
</template>

<style scoped>
.page h1 {
  margin: 0 0 12px;
  font-size: 1.5rem;
}

.page p {
  margin: 0 0 8px;
  color: #4b5563;
}

.page__meta {
  margin-top: 16px;
  padding: 12px 16px;
  background: #f3f4f6;
  border-radius: 8px;
}
</style>
