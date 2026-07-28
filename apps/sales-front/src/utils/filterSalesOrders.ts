import type { SalesOrder } from '@/types/salesOrder'

export type SalesOrderListFilter = {
  /** Exact match against any of these order numbers (trimmed, empty ignored) */
  orderNos?: string[]
  /** Inclusive YYYY-MM-DD */
  dateFrom?: string
  /** Inclusive YYYY-MM-DD */
  dateTo?: string
  /** Substring match on customerName (trim; empty = no filter) */
  customerName?: string
}

/** Split pasted multi-value order nos: comma / Chinese comma / whitespace */
export function parseOrderNos(raw: string | string[] | undefined): string[] {
  if (raw == null) return []
  const parts = Array.isArray(raw) ? raw : [raw]
  const set = new Set<string>()
  for (const part of parts) {
    for (const token of part.split(/[,，\s]+/)) {
      const t = token.trim()
      if (t) set.add(t)
    }
  }
  return [...set]
}

export function filterSalesOrders(
  orders: SalesOrder[],
  filter: SalesOrderListFilter,
): SalesOrder[] {
  const orderNos = (filter.orderNos ?? []).map((n) => n.trim()).filter(Boolean)
  const customer = filter.customerName?.trim() ?? ''
  const dateFrom = filter.dateFrom?.trim() || undefined
  const dateTo = filter.dateTo?.trim() || undefined

  return orders.filter((order) => {
    if (orderNos.length > 0 && !orderNos.includes(order.orderNo)) {
      return false
    }
    if (dateFrom && order.orderDate < dateFrom) {
      return false
    }
    if (dateTo && order.orderDate > dateTo) {
      return false
    }
    if (customer && !order.customerName.includes(customer)) {
      return false
    }
    return true
  })
}
