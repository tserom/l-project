import { describe, expect, it } from 'vitest'
import type { SalesOrder } from '@/types/salesOrder'
import {
  buildSalesOrderExportFilename,
  salesOrdersToSheetRows,
} from './exportSalesOrdersXlsx'

const sample: SalesOrder = {
  id: '1',
  orderNo: '202608109411',
  orderDate: '2026-08-10',
  customerName: '客户A',
  warehouseName: '一号仓',
  deliveryType: '销售出库',
  lines: [],
  totalQuantity: 784,
  totalAmount: 5409.6,
  createdAt: '',
  updatedAt: '',
}

describe('exportSalesOrdersXlsx', () => {
  it('maps header and one data row', () => {
    expect(salesOrdersToSheetRows([sample])).toEqual([
      ['单据号', '日期', '客户', '仓库', '出库类型', '数量合计', '金额合计'],
      [
        '202608109411',
        '2026年8月10日',
        '客户A',
        '一号仓',
        '销售出库',
        '784.00',
        '5,409.60',
      ],
    ])
  })

  it('builds filename', () => {
    expect(buildSalesOrderExportFilename(new Date('2026-08-10T15:30:45'))).toBe(
      '销售单_20260810_153045.xlsx',
    )
  })
})
