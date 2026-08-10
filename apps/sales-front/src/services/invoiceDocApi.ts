/**
 * 入口索引
 * | 动作 | 函数 | HTTP |
 * | 候选 | listInvoiceCandidates | GET /api/v1/sale/orders/invoice-candidates |
 * | 列表 | listInvoiceDocs | GET /api/v1/sale/invoice-docs |
 * | 详情 | getInvoiceDoc | GET /api/v1/sale/invoice-docs/:id |
 * | 创建 | createInvoiceDoc | POST /api/v1/sale/invoice-docs |
 * | 作废 | voidInvoiceDoc | POST /api/v1/sale/invoice-docs/:id/void |
 */
import type {
  InvoiceCandidateFilter,
  InvoiceCandidateLine,
  InvoiceDoc,
  InvoiceDocLine,
} from '@/types/invoiceDoc'
import type { SalesOrder } from '@/types/salesOrder'
import { parseOrderNos } from '@/utils/filterSalesOrders'
import { request, withQuery, type PageResult } from './httpClient'

const ORDER_BASE = '/api/v1/sale/orders'
const INVOICE_BASE = '/api/v1/sale/invoice-docs'

function isPendingInvoiceLine(line: SalesOrder['lines'][number]): boolean {
  return Boolean(line.needInvoice) && !line.invoiceDocId
}

/** Pure helper for unit tests / offline preview (does not call API). */
export function listInvoiceCandidatesFromOrders(
  orders: SalesOrder[],
  filter: InvoiceCandidateFilter,
): InvoiceCandidateLine[] {
  const orderNos = parseOrderNos(filter.orderNos)
  const customer = filter.customerName?.trim() ?? ''
  const dateFrom = filter.dateFrom?.trim() || undefined
  const dateTo = filter.dateTo?.trim() || undefined
  const out: InvoiceCandidateLine[] = []

  for (const order of orders) {
    if (customer && !order.customerName.includes(customer)) continue
    if (dateFrom && order.orderDate < dateFrom) continue
    if (dateTo && order.orderDate > dateTo) continue
    if (orderNos.length > 0 && !orderNos.includes(order.orderNo)) continue

    for (const line of order.lines) {
      if (!isPendingInvoiceLine(line)) continue
      out.push({
        id: crypto.randomUUID(),
        salesOrderId: order.id,
        salesOrderLineId: line.id,
        materialName: line.materialName,
        spec: line.spec,
        unit: line.unit,
        quantity: line.quantity,
        unitPrice: line.unitPrice,
        amount: line.amount,
        lineRemark: line.lineRemark,
        orderNo: order.orderNo,
        orderDate: order.orderDate,
        customerName: order.customerName,
        warehouseName: order.warehouseName,
        deliveryType: order.deliveryType,
      })
    }
  }
  return out
}

export async function listInvoiceCandidates(
  filter: InvoiceCandidateFilter,
): Promise<InvoiceCandidateLine[]> {
  const orderNos = parseOrderNos(filter.orderNos)
  const path = withQuery(`${ORDER_BASE}/invoice-candidates`, {
    'qp-customerName-like': filter.customerName?.trim() || undefined,
    'qp-orderNo-in': orderNos.length ? orderNos.join(',') : undefined,
    'qp-orderDate-gte': filter.dateFrom?.trim() || undefined,
    'qp-orderDate-lte': filter.dateTo?.trim() || undefined,
  })
  return request<InvoiceCandidateLine[]>(path)
}

export async function listInvoiceDocs(opts?: {
  status?: string
  page?: number
  pageSize?: number
}): Promise<PageResult<InvoiceDoc>> {
  const path = withQuery(INVOICE_BASE, {
    'qp-status-eq': opts?.status,
    page: opts?.page ?? 1,
    pageSize: opts?.pageSize ?? 100,
  })
  const data = await request<PageResult<InvoiceDoc>>(path)
  return { ...data, list: data.list ?? [] }
}

export async function getInvoiceDoc(id: string): Promise<InvoiceDoc> {
  return request<InvoiceDoc>(`${INVOICE_BASE}/${id}`)
}

export type CreateInvoiceDocInput = {
  filterCustomerName?: string
  filterDateFrom?: string
  filterDateTo?: string
  filterOrderNos?: string[]
  lines: InvoiceDocLine[]
}

export async function createInvoiceDoc(
  input: CreateInvoiceDocInput,
): Promise<InvoiceDoc> {
  return request<InvoiceDoc>(INVOICE_BASE, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function voidInvoiceDoc(id: string): Promise<InvoiceDoc> {
  return request<InvoiceDoc>(`${INVOICE_BASE}/${id}/void`, { method: 'POST' })
}
