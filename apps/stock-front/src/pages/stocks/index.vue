<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { listMaterials, listStocks, type Material, type StockBalance } from '@/api/manage'
import '@/styles/page.css'

const stocks = ref<StockBalance[]>([])
const materials = ref<Material[]>([])
const gradeFilter = ref('')
const loading = ref(false)
const error = ref('')

const materialMap = computed(() => new Map(materials.value.map((m) => [m.id, m])))

const filteredStocks = computed(() => {
  if (!gradeFilter.value.trim()) return stocks.value
  const grade = gradeFilter.value.trim()
  const ids = new Set(
    materials.value.filter((m) => m.grade === grade).map((m) => m.id),
  )
  return stocks.value.filter((s) => ids.has(s.materialId))
})

function materialLabel(materialId: number): string {
  const m = materialMap.value.get(materialId)
  return m ? `${m.materialCode} (${m.grade})` : String(materialId)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const matParams: Record<string, string | number> = { page: 1, pageSize: 500 }
    if (gradeFilter.value.trim()) {
      matParams['qp-grade-eq'] = gradeFilter.value.trim()
    }
    const [matRes, stockRes] = await Promise.all([
      listMaterials(matParams),
      listStocks({ page: 1, pageSize: 500 }),
    ])
    materials.value = matRes.list
    stocks.value = stockRes.list
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="page">
    <h1>库存查询</h1>

    <div class="page__toolbar">
      <div class="field">
        <label>牌号筛选 (qp-grade-eq)</label>
        <input v-model="gradeFilter" placeholder="如 304" @keyup.enter="load" />
      </div>
      <button class="btn btn--primary" type="button" :disabled="loading" @click="load">查询</button>
    </div>

    <p v-if="error" class="page__error">{{ error }}</p>

    <table class="data-table">
      <thead>
        <tr>
          <th>物料</th>
          <th>批次 ID</th>
          <th>仓库</th>
          <th>重量 (kg)</th>
          <th>数量</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="5">加载中…</td>
        </tr>
        <tr v-else-if="filteredStocks.length === 0">
          <td colspan="5">暂无数据</td>
        </tr>
        <tr v-for="s in filteredStocks" :key="s.id">
          <td>{{ materialLabel(s.materialId) }}</td>
          <td>{{ s.batchId }}</td>
          <td>{{ s.warehouse }}</td>
          <td>{{ s.weightKg }}</td>
          <td>{{ s.quantity }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>
