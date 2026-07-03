<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  confirmProcessingOrder,
  createProcessingOrder,
  getProcessingOrder,
  listProcessingOrders,
  type ProcessingFinishLine,
  type ProcessingOrder,
  type ProcessingPickLine,
} from '@/api/manage'
import '@/styles/page.css'

const orders = ref<ProcessingOrder[]>([])
const selected = ref<ProcessingOrder | null>(null)
const loading = ref(false)
const error = ref('')
const showModal = ref(false)

const form = reactive({
  operator: 'admin',
  remark: '',
  pickLines: [{ materialId: 0, batchId: 0, warehouse: 'DEFAULT', weightKg: '0' }] as ProcessingPickLine[],
  finishLines: [{ materialId: 0, batchId: 0, warehouse: 'DEFAULT', quantity: '0', weightKg: '' }] as ProcessingFinishLine[],
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await listProcessingOrders({ page: 1, pageSize: 50 })
    orders.value = res.list
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function viewDetail(id: number) {
  try {
    selected.value = await getProcessingOrder(id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载详情失败'
  }
}

function addPickLine() {
  form.pickLines.push({ materialId: 0, batchId: 0, warehouse: 'DEFAULT', weightKg: '0' })
}

function addFinishLine() {
  form.finishLines.push({ materialId: 0, batchId: 0, warehouse: 'DEFAULT', quantity: '0', weightKg: '' })
}

async function submit() {
  error.value = ''
  try {
    const pickLines = form.pickLines.map((l) => ({
      ...l,
      materialId: Number(l.materialId),
      batchId: Number(l.batchId),
    }))
    const finishLines = form.finishLines.map((l) => ({
      ...l,
      materialId: Number(l.materialId),
      batchId: Number(l.batchId),
      weightKg: l.weightKg?.trim() ? l.weightKg : undefined,
    }))
    await createProcessingOrder({
      operator: form.operator,
      remark: form.remark,
      pickLines,
      finishLines,
    })
    showModal.value = false
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '创建失败'
  }
}

async function confirm(id: number) {
  error.value = ''
  try {
    await confirmProcessingOrder(id)
    await load()
    await viewDetail(id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '确认失败'
  }
}

onMounted(load)
</script>

<template>
  <section class="page">
    <h1>加工单</h1>

    <div class="page__toolbar">
      <button class="btn btn--primary" type="button" @click="showModal = true">新建加工单</button>
      <button class="btn" type="button" :disabled="loading" @click="load">刷新</button>
    </div>

    <p v-if="error" class="page__error">{{ error }}</p>

    <table class="data-table">
      <thead>
        <tr>
          <th>单号</th>
          <th>日期</th>
          <th>状态</th>
          <th>损耗 (kg)</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="5">加载中…</td>
        </tr>
        <tr v-else-if="orders.length === 0">
          <td colspan="5">暂无数据</td>
        </tr>
        <tr v-for="o in orders" :key="o.id">
          <td>{{ o.docNo }}</td>
          <td>{{ o.docDate?.slice(0, 10) }}</td>
          <td>
            <span class="status" :class="o.status === 'confirmed' ? 'status--confirmed' : 'status--draft'">
              {{ o.status }}
            </span>
          </td>
          <td>{{ o.status === 'confirmed' ? o.lossWeightKg : '—' }}</td>
          <td>
            <button class="btn btn--sm" type="button" @click="viewDetail(o.id)">详情</button>
            <button
              v-if="o.status === 'draft'"
              class="btn btn--sm btn--primary"
              type="button"
              @click="confirm(o.id)"
            >
              确认
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="selected" class="detail-panel">
      <h3>{{ selected.docNo }}</h3>
      <p v-if="selected.status === 'confirmed'" style="margin: 0 0 12px">
        <strong>损耗重量：</strong>{{ selected.lossWeightKg }} kg
      </p>

      <h4>领料行</h4>
      <table class="data-table">
        <thead>
          <tr>
            <th>物料 ID</th>
            <th>批次 ID</th>
            <th>仓库</th>
            <th>重量 (kg)</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(line, i) in selected.pickLines ?? []" :key="'p' + i">
            <td>{{ line.materialId }}</td>
            <td>{{ line.batchId }}</td>
            <td>{{ line.warehouse }}</td>
            <td>{{ line.weightKg }}</td>
          </tr>
        </tbody>
      </table>

      <h4 style="margin-top: 16px">完工行</h4>
      <table class="data-table">
        <thead>
          <tr>
            <th>物料 ID</th>
            <th>批次 ID</th>
            <th>仓库</th>
            <th>数量</th>
            <th>重量 (kg)</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(line, i) in selected.finishLines ?? []" :key="'f' + i">
            <td>{{ line.materialId }}</td>
            <td>{{ line.batchId }}</td>
            <td>{{ line.warehouse }}</td>
            <td>{{ line.quantity }}</td>
            <td>{{ line.weightKg ?? '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
      <div class="modal">
        <h2>新建加工单</h2>
        <div class="form-grid">
          <div class="field">
            <label>经办人</label>
            <input v-model="form.operator" />
          </div>
          <div class="field">
            <label>备注</label>
            <input v-model="form.remark" />
          </div>
        </div>

        <h4>领料行</h4>
        <div v-for="(line, i) in form.pickLines" :key="'p' + i" class="line-block">
          <p class="line-block__title">领料 {{ i + 1 }}</p>
          <div class="form-grid">
            <div class="field">
              <label>物料 ID</label>
              <input v-model.number="line.materialId" type="number" />
            </div>
            <div class="field">
              <label>批次 ID</label>
              <input v-model.number="line.batchId" type="number" />
            </div>
            <div class="field">
              <label>仓库</label>
              <input v-model="line.warehouse" />
            </div>
            <div class="field">
              <label>重量 (kg)</label>
              <input v-model="line.weightKg" />
            </div>
          </div>
        </div>
        <button class="btn btn--sm" type="button" @click="addPickLine">+ 添加领料</button>

        <h4 style="margin-top: 16px">完工行</h4>
        <div v-for="(line, i) in form.finishLines" :key="'f' + i" class="line-block">
          <p class="line-block__title">完工 {{ i + 1 }}</p>
          <div class="form-grid">
            <div class="field">
              <label>物料 ID</label>
              <input v-model.number="line.materialId" type="number" />
            </div>
            <div class="field">
              <label>批次 ID</label>
              <input v-model.number="line.batchId" type="number" />
            </div>
            <div class="field">
              <label>仓库</label>
              <input v-model="line.warehouse" />
            </div>
            <div class="field">
              <label>数量</label>
              <input v-model="line.quantity" />
            </div>
            <div class="field">
              <label>重量 (kg，可选)</label>
              <input v-model="line.weightKg" />
            </div>
          </div>
        </div>
        <button class="btn btn--sm" type="button" @click="addFinishLine">+ 添加完工</button>

        <div class="modal__actions">
          <button class="btn" type="button" @click="showModal = false">取消</button>
          <button class="btn btn--primary" type="button" @click="submit">保存</button>
        </div>
      </div>
    </div>
  </section>
</template>
