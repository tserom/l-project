import * as XLSX from 'xlsx'
import type { SalesOrder } from '@/types/salesOrder'
import { formatOrderDate } from '@/utils/dateFormat'
import { formatAmount, formatQuantity } from '@/utils/money'

const HEADER = [
  '单据号',
  '日期',
  '客户',
  '仓库',
  '出库类型',
  '数量合计',
  '金额合计',
] as const

export function salesOrdersToSheetRows(orders: SalesOrder[]): string[][] {
  return [
    [...HEADER],
    ...orders.map((o) => [
      o.orderNo,
      formatOrderDate(o.orderDate),
      o.customerName,
      o.warehouseName,
      o.deliveryType,
      formatQuantity(o.totalQuantity),
      formatAmount(o.totalAmount),
    ]),
  ]
}

export function buildSalesOrderExportFilename(now: Date = new Date()): string {
  const p = (n: number) => String(n).padStart(2, '0')
  const stamp =
    `${now.getFullYear()}${p(now.getMonth() + 1)}${p(now.getDate())}` +
    `_${p(now.getHours())}${p(now.getMinutes())}${p(now.getSeconds())}`
  return `销售单_${stamp}.xlsx`
}

export function downloadSalesOrdersXlsx(
  orders: SalesOrder[],
  now: Date = new Date(),
): void {
  const rows = salesOrdersToSheetRows(orders)
  const sheet = XLSX.utils.aoa_to_sheet(rows)
  const book = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(book, sheet, '销售单')
  XLSX.writeFile(book, buildSalesOrderExportFilename(now))
}
