import { describe, expect, it } from 'vitest'
import type { SalesOrder } from '@/types/salesOrder'
import { filterSalesOrders, parseOrderNos } from './filterSalesOrders'

const base: SalesOrder = {
  id: '1',
  orderNo: '00150262',
  orderDate: '2026-07-22',
  customerName: '884周村 马俊生',
  warehouseName: '01金阳仓库',
  deliveryType: '提货',
  lines: [],
  totalQuantity: 1,
  totalAmount: 1,
  createdAt: '',
  updatedAt: '',
}

const orders: SalesOrder[] = [
  base,
  {
    ...base,
    id: '2',
    orderNo: '202607247712',
    orderDate: '2026-07-24',
    customerName: '啊实打实的',
  },
]

describe('parseOrderNos', () => {
  it('splits comma and whitespace', () => {
    expect(parseOrderNos('a, b，c\nd')).toEqual(['a', 'b', 'c', 'd'])
  })
})

describe('filterSalesOrders', () => {
  it('filters by multi orderNos', () => {
    expect(
      filterSalesOrders(orders, { orderNos: ['202607247712', 'x'] }).map((o) => o.orderNo),
    ).toEqual(['202607247712'])
  })

  it('filters by date range inclusive', () => {
    expect(
      filterSalesOrders(orders, { dateFrom: '2026-07-22', dateTo: '2026-07-22' }),
    ).toHaveLength(1)
    expect(
      filterSalesOrders(orders, { dateFrom: '2026-07-22', dateTo: '2026-07-24' }),
    ).toHaveLength(2)
  })

  it('filters customer by substring', () => {
    expect(filterSalesOrders(orders, { customerName: '马俊' })).toHaveLength(1)
    expect(filterSalesOrders(orders, { customerName: '  ' })).toHaveLength(2)
  })
})
