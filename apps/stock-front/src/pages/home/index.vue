<script setup lang="ts">
import { RouterLink } from 'vue-router'
import '@/styles/page.css'

const links = [
  { to: '/materials', title: '物料档案', desc: '维护材质、规格与编码' },
  { to: '/stocks', title: '库存查询', desc: '按牌号查看库存余额' },
  { to: '/inbound', title: '入库单', desc: '来料与完工入库' },
  { to: '/outbound', title: '出库单', desc: '领料与其它出库' },
  { to: '/sales', title: '销售单', desc: '销售下单与出库' },
  { to: '/processing', title: '加工单', desc: '领料加工与损耗' },
]
</script>

<template>
  <section class="page">
    <h1>不锈钢进销存</h1>
    <p style="margin: 0 0 20px; color: #6b7280">独立前端，通过 stock-manage API 管理库存与单据。</p>

    <div class="link-grid">
      <RouterLink v-for="item in links" :key="item.to" :to="item.to" class="link-card">
        <strong>{{ item.title }}</strong>
        <span>{{ item.desc }}</span>
      </RouterLink>
    </div>
  </section>
</template>
