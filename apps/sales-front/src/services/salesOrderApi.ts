/**
 * 入口索引
 * | 动作 | 函数 | HTTP |
 * | 列表 | listSalesOrders | GET /api/v1/sale/orders |
 * | 详情 | getSalesOrder | GET /api/v1/sale/orders/:id |
 * | 新建 | createSalesOrder | POST /api/v1/sale/orders |
 * | 更新 | updateSalesOrder | PUT /api/v1/sale/orders/:id |
 * | 删除 | removeSalesOrder | DELETE /api/v1/sale/orders/:id |
 * | 批量导出 | exportSalesOrdersToExcel | GET /api/v1/sale/orders（可能多次；前端写 xlsx） |
 */
import type { SalesOrder, SalesOrderInput } from '@/types/salesOrder'
import { downloadSalesOrdersXlsx } from '@/utils/exportSalesOrdersXlsx'
import { request, withQuery, type PageResult } from './httpClient'

const BASE = '/api/v1/sale/orders'

export type ListSalesOrdersQuery = {
  orderNos?: string[]
  dateFrom?: string
  dateTo?: string
  customerName?: string
  page?: number
  pageSize?: number
}

function hydrateOrder(order: SalesOrder): SalesOrder {
  return {
    ...order,
    lines: (order.lines ?? []).map((line) => ({
      ...line,
      needInvoice: Boolean(line.needInvoice),
    })),
  }
}

export async function listSalesOrders(
  query: ListSalesOrdersQuery = {},
): Promise<PageResult<SalesOrder>> {
  const orderNos = (query.orderNos ?? []).map((n) => n.trim()).filter(Boolean)
  const path = withQuery(BASE, {
    'qp-customerName-like': query.customerName?.trim() || undefined,
    'qp-orderNo-in': orderNos.length ? orderNos.join(',') : undefined,
    'qp-orderDate-gte': query.dateFrom?.trim() || undefined,
    'qp-orderDate-lte': query.dateTo?.trim() || undefined,
    page: query.page ?? 1,
    pageSize: query.pageSize ?? 100,
  })
  const data = await request<PageResult<SalesOrder>>(path)
  return {
    ...data,
    list: (data.list ?? []).map(hydrateOrder),
  }
}

export async function getSalesOrder(id: string): Promise<SalesOrder> {
  return hydrateOrder(await request<SalesOrder>(`${BASE}/${id}`))
}

export async function createSalesOrder(input: SalesOrderInput): Promise<SalesOrder> {
  return hydrateOrder(
    await request<SalesOrder>(BASE, {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  )
}

export async function updateSalesOrder(
  id: string,
  input: SalesOrderInput,
): Promise<SalesOrder> {
  return hydrateOrder(
    await request<SalesOrder>(`${BASE}/${id}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  )
}

export async function removeSalesOrder(id: string): Promise<void> {
  await request<{ ok: boolean }>(`${BASE}/${id}`, { method: 'DELETE' })
}

export type ExportSalesOrdersResult =
  | { ok: true }
  | { ok: false; reason: 'empty' }

async function listAllSalesOrdersForExport(
  query: ListSalesOrdersQuery = {},
): Promise<SalesOrder[]> {
  const pageSize = 500
  let page = 1
  const all: SalesOrder[] = []
  let total = Infinity
  while (all.length < total) {
    const result = await listSalesOrders({ ...query, page, pageSize })
    total = result.total
    if (!result.list.length) break
    all.push(...result.list)
    page += 1
  }
  return all
}

/** L3：勾选优先；否则按筛选分页拉全，前端生成汇总 xlsx */
export async function exportSalesOrdersToExcel(input: {
  selectedOrders?: SalesOrder[]
  query?: ListSalesOrdersQuery
}): Promise<ExportSalesOrdersResult> {
  const selected = input.selectedOrders ?? []
  const rows =
    selected.length > 0
      ? selected
      : await listAllSalesOrdersForExport(input.query ?? {})
  if (rows.length === 0) return { ok: false, reason: 'empty' }
  downloadSalesOrdersXlsx(rows)
  return { ok: true }
}
