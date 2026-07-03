<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  confirmSalesOrder,
  confirmSalesShipment,
  createSalesOrder,
  createSalesShipment,
  getSalesOrder,
  listSalesOrders,
  listSalesShipments,
  type SalesLine,
  type SalesOrder,
  type SalesShipment,
} from '@/api/manage'
import '@/styles/page.css'

const orders = ref<SalesOrder[]>([])
const selected = ref<SalesOrder | null>(null)
const shipments = ref<SalesShipment[]>([])
const loading = ref(false)
const error = ref('')
const success = ref('')
const showModal = ref(false)
const shipmentWarehouse = ref('DEFAULT')

const form = reactive({
  customerName: '',
  operator: 'admin',
  remark: '',
  lines: [{ materialId: 0, batchId: 0, weightKg: '0', quantity: '0' }] as SalesLine[],
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await listSalesOrders({ page: 1, pageSize: 50 })
    orders.value = res.list
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function viewDetail(id: number) {
  success.value = ''
  try {
    selected.value = await getSalesOrder(id)
    shipments.value = await listSalesShipments(id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载详情失败'
  }
}

function addLine() {
  form.lines.push({ materialId: 0, batchId: 0, weightKg: '0', quantity: '0' })
}

async function submit() {
  error.value = ''
  try {
    const lines = form.lines.map((l) => ({
      ...l,
      materialId: Number(l.materialId),
      batchId: Number(l.batchId),
    }))
    await createSalesOrder({
      customerName: form.customerName,
      operator: form.operator,
      remark: form.remark,
      lines,
    })
    showModal.value = false
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '创建失败'
  }
}

async function confirm(id: number) {
  error.value = ''
  success.value = ''
  try {
    await confirmSalesOrder(id)
    await load()
    if (selected.value?.id === id) await viewDetail(id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '确认失败'
  }
}

async function createShipment(salesOrderId: number) {
  error.value = ''
  success.value = ''
  try {
    const shipment = await createSalesShipment(salesOrderId, {
      operator: 'admin',
      warehouse: shipmentWarehouse.value,
    })
    success.value = `已创建出库单 ${shipment.docNo}`
    await viewDetail(salesOrderId)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '创建出库单失败'
  }
}

async function confirmShipment(id: number, salesOrderId: number) {
  error.value = ''
  success.value = ''
  try {
    await confirmSalesShipment(id)
    success.value = '出库单已确认'
    await viewDetail(salesOrderId)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '确认出库失败'
  }
}

onMounted(load)
</script>

<template>
  <section class="page">
    <h1>销售单</h1>

    <div class="page__toolbar">
      <button class="btn btn--primary" type="button" @click="showModal = true">新建销售单</button>
      <button class="btn" type="button" :disabled="loading" @click="load">刷新</button>
    </div>

    <p v-if="error" class="page__error">{{ error }}</p>
    <p v-if="success" class="page__success">{{ success }}</p>

    <table class="data-table">
      <thead>
        <tr>
          <th>单号</th>
          <th>客户</th>
          <th>日期</th>
          <th>状态</th>
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
          <td>{{ o.customerName }}</td>
          <td>{{ o.docDate?.slice(0, 10) }}</td>
          <td>
            <span class="status" :class="o.status === 'confirmed' ? 'status--confirmed' : 'status--draft'">
              {{ o.status }}
            </span>
          </td>
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
      <h3>{{ selected.docNo }} — {{ selected.customerName }}</h3>
      <table class="data-table">
        <thead>
          <tr>
            <th>物料 ID</th>
            <th>批次 ID</th>
            <th>重量</th>
            <th>数量</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(line, i) in selected.lines ?? []" :key="i">
            <td>{{ line.materialId }}</td>
            <td>{{ line.batchId }}</td>
            <td>{{ line.weightKg }}</td>
            <td>{{ line.quantity }}</td>
          </tr>
        </tbody>
      </table>

      <div v-if="selected.status === 'confirmed'" class="page__toolbar" style="margin-top: 16px">
        <div class="field">
          <label>出库仓库</label>
          <input v-model="shipmentWarehouse" />
        </div>
        <button class="btn btn--primary btn--sm" type="button" @click="createShipment(selected.id)">
          创建出库单
        </button>
      </div>

      <div v-if="shipments.length" style="margin-top: 16px">
        <h4>销售出库单</h4>
        <table class="data-table">
          <thead>
            <tr>
              <th>单号</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in shipments" :key="s.id">
              <td>{{ s.docNo }}</td>
              <td>
                <span class="status" :class="s.status === 'confirmed' ? 'status--confirmed' : 'status--draft'">
                  {{ s.status }}
                </span>
              </td>
              <td>
                <button
                  v-if="s.status === 'draft'"
                  class="btn btn--sm btn--primary"
                  type="button"
                  @click="confirmShipment(s.id, selected!.id)"
                >
                  确认出库
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
      <div class="modal">
        <h2>新建销售单</h2>
        <div class="form-grid">
          <div class="field">
            <label>客户名称</label>
            <input v-model="form.customerName" />
          </div>
          <div class="field">
            <label>经办人</label>
            <input v-model="form.operator" />
          </div>
          <div class="field">
            <label>备注</label>
            <input v-model="form.remark" />
          </div>
        </div>
        <div v-for="(line, i) in form.lines" :key="i" class="line-block">
          <p class="line-block__title">明细 {{ i + 1 }}</p>
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
              <label>重量 (kg)</label>
              <input v-model="line.weightKg" />
            </div>
            <div class="field">
              <label>数量</label>
              <input v-model="line.quantity" />
            </div>
          </div>
        </div>
        <button class="btn btn--sm" type="button" @click="addLine">+ 添加明细</button>
        <div class="modal__actions">
          <button class="btn" type="button" @click="showModal = false">取消</button>
          <button class="btn btn--primary" type="button" @click="submit">保存</button>
        </div>
      </div>
    </div>
  </section>
</template>
