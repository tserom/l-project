<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  createMaterial,
  listMaterials,
  type CreateMaterialInput,
  type Material,
} from '@/api/manage'
import '@/styles/page.css'

const materials = ref<Material[]>([])
const loading = ref(false)
const error = ref('')
const showModal = ref(false)

const form = reactive<CreateMaterialInput>({
  materialCode: '',
  grade: '304',
  form: 'plate',
  spec: '',
  primaryUnit: 'kg',
  materialType: 'raw',
})

const formOptions = {
  forms: [
    { value: 'plate', label: '板' },
    { value: 'pipe', label: '管' },
    { value: 'bar', label: '棒' },
    { value: 'profile', label: '型材' },
    { value: 'part', label: '零件' },
  ],
  units: [
    { value: 'kg', label: 'kg' },
    { value: 'piece', label: '件' },
    { value: 'meter', label: '米' },
  ],
  types: [
    { value: 'raw', label: '原材料' },
    { value: 'finished', label: '成品' },
  ],
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await listMaterials({ page: 1, pageSize: 100 })
    materials.value = res.list
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.materialCode = ''
  form.grade = '304'
  form.form = 'plate'
  form.spec = ''
  form.primaryUnit = 'kg'
  form.materialType = 'raw'
  showModal.value = true
}

async function submit() {
  error.value = ''
  try {
    await createMaterial({ ...form })
    showModal.value = false
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '创建失败'
  }
}

onMounted(load)
</script>

<template>
  <section class="page">
    <h1>物料档案</h1>

    <div class="page__toolbar">
      <button class="btn btn--primary" type="button" @click="openCreate">新建物料</button>
      <button class="btn" type="button" :disabled="loading" @click="load">刷新</button>
    </div>

    <p v-if="error" class="page__error">{{ error }}</p>

    <table class="data-table">
      <thead>
        <tr>
          <th>编码</th>
          <th>牌号</th>
          <th>形态</th>
          <th>规格</th>
          <th>主单位</th>
          <th>类型</th>
          <th>状态</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="7">加载中…</td>
        </tr>
        <tr v-else-if="materials.length === 0">
          <td colspan="7">暂无数据</td>
        </tr>
        <tr v-for="m in materials" :key="m.id">
          <td>{{ m.materialCode }}</td>
          <td>{{ m.grade }}</td>
          <td>{{ m.form }}</td>
          <td>{{ m.spec }}</td>
          <td>{{ m.primaryUnit }}</td>
          <td>{{ m.materialType }}</td>
          <td>{{ m.status }}</td>
        </tr>
      </tbody>
    </table>

    <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
      <div class="modal">
        <h2>新建物料</h2>
        <div class="form-grid">
          <div class="field">
            <label>物料编码</label>
            <input v-model="form.materialCode" required />
          </div>
          <div class="field">
            <label>牌号</label>
            <input v-model="form.grade" />
          </div>
          <div class="field">
            <label>形态</label>
            <select v-model="form.form">
              <option v-for="o in formOptions.forms" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
          <div class="field">
            <label>规格</label>
            <input v-model="form.spec" />
          </div>
          <div class="field">
            <label>主单位</label>
            <select v-model="form.primaryUnit">
              <option v-for="o in formOptions.units" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
          <div class="field">
            <label>物料类型</label>
            <select v-model="form.materialType">
              <option v-for="o in formOptions.types" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
        </div>
        <div class="modal__actions">
          <button class="btn" type="button" @click="showModal = false">取消</button>
          <button class="btn btn--primary" type="button" @click="submit">保存</button>
        </div>
      </div>
    </div>
  </section>
</template>
