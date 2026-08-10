import * as XLSX from 'xlsx'
import type { SalesOrder } from '@/types/salesOrder'
import { formatOrderDate } from '@/utils/dateFormat'
import { formatAmount, formatQuantity, sumAmounts } from '@/utils/money'

/** 导出列定义：key 与 SalesOrder / 列表列对齐，禁止靠空字符串数格占位 */
const EXPORT_COLUMNS = [
  { key: 'orderNo', title: '单据号' },
  { key: 'orderDate', title: '日期' },
  { key: 'customerName', title: '客户' },
  { key: 'warehouseName', title: '仓库' },
  { key: 'deliveryType', title: '出库类型' },
  { key: 'totalQuantity', title: '数量合计' },
  { key: 'totalAmount', title: '金额合计' },
] as const

type ExportColumnKey = (typeof EXPORT_COLUMNS)[number]['key']

function cellForOrder(order: SalesOrder, key: ExportColumnKey): string {
  switch (key) {
    case 'orderNo':
      return order.orderNo
    case 'orderDate':
      return formatOrderDate(order.orderDate)
    case 'customerName':
      return order.customerName
    case 'warehouseName':
      return order.warehouseName
    case 'deliveryType':
      return order.deliveryType
    case 'totalQuantity':
      return formatQuantity(order.totalQuantity)
    case 'totalAmount':
      return formatAmount(order.totalAmount)
  }
}

function cellsFromByKey(
  byKey: Partial<Record<ExportColumnKey, string>>,
): string[] {
  return EXPORT_COLUMNS.map((col) => byKey[col.key] ?? '')
}

export function salesOrdersToSheetRows(orders: SalesOrder[]): string[][] {
  const header = EXPORT_COLUMNS.map((col) => col.title)
  const dataRows = orders.map((order) =>
    EXPORT_COLUMNS.map((col) => cellForOrder(order, col.key)),
  )
  const totalAmount = sumAmounts(orders.map((o) => ({ amount: o.totalAmount })))
  const summaryRow = cellsFromByKey({
    orderNo: '合计',
    totalAmount: formatAmount(totalAmount),
  })
  return [header, ...dataRows, summaryRow]
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
