/**
 * 入口索引
 * | 动作 | 函数 | 存储 |
 * | 候选 | listInvoiceCandidates | 扫描 salesOrders |
 * | 列表 | listInvoiceDocs | invoiceDocs |
 * | 详情 | getInvoiceDoc | 同上 |
 * | 创建 | createInvoiceDoc | 写开票单 + 回写明细 invoiceDocId |
 * | 作废 | voidInvoiceDoc | 作废 + 清空明细 invoiceDocId |
 */
import { db } from '@/storage/db'
import * as invoiceRepo from '@/storage/invoiceDocRepository'
import * as orderRepo from '@/storage/salesOrderRepository'
import type {
  InvoiceCandidateFilter,
  InvoiceCandidateLine,
  InvoiceDoc,
  InvoiceDocLine,
} from '@/types/invoiceDoc'
import type { SalesOrder } from '@/types/salesOrder'
import { parseOrderNos } from '@/utils/filterSalesOrders'
import { sumAmounts, sumQuantities } from '@/utils/money'
import { generateOrderNo } from '@/utils/orderNo'

function isPendingInvoiceLine(line: SalesOrder['lines'][number]): boolean {
  return Boolean(line.needInvoice) && !line.invoiceDocId
}

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
  const orders = await orderRepo.listAll()
  return listInvoiceCandidatesFromOrders(orders, filter)
}

export async function listInvoiceDocs(): Promise<InvoiceDoc[]> {
  return invoiceRepo.listAll()
}

export async function getInvoiceDoc(id: string): Promise<InvoiceDoc> {
  const doc = await invoiceRepo.getById(id)
  if (!doc) throw new Error('开票单不存在')
  return doc
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
  if (!input.lines.length) {
    throw new Error('至少保留一行明细')
  }

  const now = new Date().toISOString()
  const docId = crypto.randomUUID()
  const lines = input.lines.map((l) => ({ ...l, id: l.id || crypto.randomUUID() }))

  await db.transaction('rw', db.salesOrders, db.invoiceDocs, async () => {
    for (const line of lines) {
      const order = await db.salesOrders.get(line.salesOrderId)
      if (!order) {
        throw new Error(`销售单不存在：${line.orderNo}`)
      }
      const src = order.lines.find((x) => x.id === line.salesOrderLineId)
      if (!src) {
        throw new Error(`明细不存在：${line.materialName}`)
      }
      if (!src.needInvoice) {
        throw new Error(`明细未勾选需开票：${line.orderNo} / ${line.materialName}`)
      }
      if (src.invoiceDocId) {
        throw new Error(`明细已开票，请重新加载：${line.orderNo} / ${line.materialName}`)
      }
      const nextLines = order.lines.map((x) =>
        x.id === line.salesOrderLineId ? { ...x, invoiceDocId: docId } : x,
      )
      await db.salesOrders.put({
        ...order,
        lines: nextLines,
        updatedAt: now,
      })
    }

    const doc: InvoiceDoc = {
      id: docId,
      invoiceNo: `KP${generateOrderNo()}`,
      status: 'saved',
      filterCustomerName: input.filterCustomerName?.trim() || undefined,
      filterDateFrom: input.filterDateFrom,
      filterDateTo: input.filterDateTo,
      filterOrderNos: input.filterOrderNos?.length
        ? [...input.filterOrderNos]
        : undefined,
      lines,
      totalQuantity: sumQuantities(lines),
      totalAmount: sumAmounts(lines),
      createdAt: now,
      updatedAt: now,
    }
    await db.invoiceDocs.put(doc)
  })

  return getInvoiceDoc(docId)
}

export async function voidInvoiceDoc(id: string): Promise<InvoiceDoc> {
  const doc = await getInvoiceDoc(id)
  if (doc.status === 'voided') {
    throw new Error('开票单已作废')
  }
  const now = new Date().toISOString()

  await db.transaction('rw', db.salesOrders, db.invoiceDocs, async () => {
    for (const line of doc.lines) {
      const order = await db.salesOrders.get(line.salesOrderId)
      if (!order) continue
      const nextLines = order.lines.map((x) => {
        if (x.id !== line.salesOrderLineId) return x
        if (x.invoiceDocId !== id) return x
        const { invoiceDocId: _removed, ...rest } = x
        return rest
      })
      await db.salesOrders.put({
        ...order,
        lines: nextLines,
        updatedAt: now,
      })
    }
    await db.invoiceDocs.put({
      ...doc,
      status: 'voided',
      updatedAt: now,
      voidedAt: now,
    })
  })

  return getInvoiceDoc(id)
}
