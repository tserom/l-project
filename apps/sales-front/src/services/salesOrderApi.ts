/**
 * 入口索引
 * | 动作 | 函数 | HTTP |
 * | 列表 | listSalesOrders | GET /api/v1/sale/orders |
 * | 详情 | getSalesOrder | GET /api/v1/sale/orders/:id |
 * | 新建 | createSalesOrder | POST /api/v1/sale/orders |
 * | 更新 | updateSalesOrder | PUT /api/v1/sale/orders/:id |
 * | 删除 | removeSalesOrder | DELETE /api/v1/sale/orders/:id |
 */
import type { SalesOrder, SalesOrderInput } from '@/types/salesOrder'
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
