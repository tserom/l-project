/** Manage BFF client — relative `/api/v1` in prod; Vite proxies `/api` in dev. */

const API_PREFIX = `${import.meta.env.VITE_API_BASE ?? ''}/api/v1`

export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface Material {
  id: number
  materialCode: string
  grade: string
  form: string
  spec: string
  primaryUnit: string
  materialType: string
  status: string
}

export interface MaterialBatch {
  id: number
  materialId: number
  heatNo: string
  remark?: string
}

export interface StockBalance {
  id: number
  materialId: number
  batchId: number
  warehouse: string
  weightKg: string
  quantity: string
}

export interface DocLine {
  materialId: number
  batchId: number
  warehouse: string
  weightKg: string
  quantity: string
}

export interface SalesLine {
  materialId: number
  batchId: number
  weightKg: string
  quantity: string
}

export interface ProcessingPickLine {
  materialId: number
  batchId: number
  warehouse: string
  weightKg: string
}

export interface ProcessingFinishLine {
  materialId: number
  batchId: number
  warehouse: string
  quantity: string
  weightKg?: string
}

export interface InboundOrder {
  id: number
  docNo: string
  docDate: string
  status: string
  operator: string
  remark?: string
  lines?: DocLine[]
}

export interface OutboundOrder {
  id: number
  docNo: string
  docDate: string
  status: string
  operator: string
  remark?: string
  lines?: DocLine[]
}

export interface SalesOrder {
  id: number
  docNo: string
  docDate: string
  status: string
  customerName: string
  operator: string
  remark?: string
  lines?: SalesLine[]
}

export interface SalesShipment {
  id: number
  docNo: string
  docDate: string
  status: string
  salesOrderId: number
  operator: string
  remark?: string
  lines?: DocLine[]
}

export interface ProcessingOrder {
  id: number
  docNo: string
  docDate: string
  status: string
  lossWeightKg: string
  operator: string
  remark?: string
  pickLines?: ProcessingPickLine[]
  finishLines?: ProcessingFinishLine[]
}

export interface CreateMaterialInput {
  materialCode: string
  grade: string
  form: string
  spec: string
  primaryUnit: string
  materialType: string
}

type QueryParams = Record<string, string | number | undefined>

function buildQuery(params?: QueryParams): string {
  if (!params) return ''
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') qs.set(key, String(value))
  }
  const s = qs.toString()
  return s ? `?${s}` : ''
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  if (init?.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(`${API_PREFIX}${path}`, { ...init, headers })
  const body = (await res.json()) as ApiResponse<T>

  if (!res.ok || body.code !== 0) {
    throw new Error(body.message || res.statusText)
  }

  return body.data as T
}

// Materials & batches (proxied to center)

export function listMaterials(params?: QueryParams) {
  return request<PageResult<Material>>(`/materials${buildQuery(params)}`)
}

export function createMaterial(input: CreateMaterialInput) {
  return request<Material>('/materials', { method: 'POST', body: JSON.stringify(input) })
}

export function listBatches(params?: QueryParams) {
  return request<PageResult<MaterialBatch>>(`/batches${buildQuery(params)}`)
}

export function createBatch(input: { materialId: number; heatNo: string; remark?: string }) {
  return request<MaterialBatch>('/batches', { method: 'POST', body: JSON.stringify(input) })
}

// Stocks (proxied to center)

export function listStocks(params?: QueryParams) {
  return request<PageResult<StockBalance>>(`/stocks${buildQuery(params)}`)
}

// Inbound orders

export function listInboundOrders(params?: QueryParams) {
  return request<PageResult<InboundOrder>>(`/inbound-orders${buildQuery(params)}`)
}

export function getInboundOrder(id: number) {
  return request<InboundOrder>(`/inbound-orders/${id}`)
}

export function createInboundOrder(input: {
  docDate?: string
  operator: string
  remark?: string
  lines: DocLine[]
}) {
  return request<InboundOrder>('/inbound-orders', { method: 'POST', body: JSON.stringify(input) })
}

export function confirmInboundOrder(id: number) {
  return request<InboundOrder>(`/inbound-orders/${id}/confirm`, { method: 'POST' })
}

// Outbound orders

export function listOutboundOrders(params?: QueryParams) {
  return request<PageResult<OutboundOrder>>(`/outbound-orders${buildQuery(params)}`)
}

export function getOutboundOrder(id: number) {
  return request<OutboundOrder>(`/outbound-orders/${id}`)
}

export function createOutboundOrder(input: {
  docDate?: string
  operator: string
  remark?: string
  lines: DocLine[]
}) {
  return request<OutboundOrder>('/outbound-orders', { method: 'POST', body: JSON.stringify(input) })
}

export function confirmOutboundOrder(id: number) {
  return request<OutboundOrder>(`/outbound-orders/${id}/confirm`, { method: 'POST' })
}

// Sales orders

export function listSalesOrders(params?: QueryParams) {
  return request<PageResult<SalesOrder>>(`/sales-orders${buildQuery(params)}`)
}

export function getSalesOrder(id: number) {
  return request<SalesOrder>(`/sales-orders/${id}`)
}

export function createSalesOrder(input: {
  docDate?: string
  customerName: string
  operator: string
  remark?: string
  lines: SalesLine[]
}) {
  return request<SalesOrder>('/sales-orders', { method: 'POST', body: JSON.stringify(input) })
}

export function confirmSalesOrder(id: number) {
  return request<SalesOrder>(`/sales-orders/${id}/confirm`, { method: 'POST' })
}

export function createSalesShipment(
  salesOrderId: number,
  input: { docDate?: string; operator: string; remark?: string; warehouse: string },
) {
  return request<SalesShipment>(`/sales-orders/${salesOrderId}/shipments`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function confirmSalesShipment(id: number) {
  return request<SalesShipment>(`/sales-shipments/${id}/confirm`, { method: 'POST' })
}

export function listSalesShipments(salesOrderId: number) {
  return request<SalesShipment[]>(`/sales-orders/${salesOrderId}/shipments`)
}

// Processing orders

export function listProcessingOrders(params?: QueryParams) {
  return request<PageResult<ProcessingOrder>>(`/processing-orders${buildQuery(params)}`)
}

export function getProcessingOrder(id: number) {
  return request<ProcessingOrder>(`/processing-orders/${id}`)
}

export function createProcessingOrder(input: {
  docDate?: string
  operator: string
  remark?: string
  pickLines: ProcessingPickLine[]
  finishLines: ProcessingFinishLine[]
}) {
  return request<ProcessingOrder>('/processing-orders', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function confirmProcessingOrder(id: number) {
  return request<ProcessingOrder>(`/processing-orders/${id}/confirm`, { method: 'POST' })
}
