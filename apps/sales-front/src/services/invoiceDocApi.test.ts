import { describe, expect, it } from 'vitest'
import { listInvoiceCandidatesFromOrders } from './invoiceDocApi'
import type { SalesOrder } from '@/types/salesOrder'

describe('listInvoiceCandidatesFromOrders', () => {
  it('only pending needInvoice lines', () => {
    const orders: SalesOrder[] = [
      {
        id: 'o1',
        orderNo: 'A1',
        orderDate: '2026-07-01',
        customerName: '客户甲',
        warehouseName: '仓',
        deliveryType: '提货',
        lines: [
          {
            id: 'l1',
            materialName: '要开',
            spec: '',
            unit: 'kg',
            quantity: 1,
            unitPrice: 2,
            amount: 2,
            needInvoice: true,
          },
          {
            id: 'l2',
            materialName: '不要',
            spec: '',
            unit: 'kg',
            quantity: 1,
            unitPrice: 2,
            amount: 2,
            needInvoice: false,
          },
          {
            id: 'l3',
            materialName: '已开',
            spec: '',
            unit: 'kg',
            quantity: 1,
            unitPrice: 2,
            amount: 2,
            needInvoice: true,
            invoiceDocId: 'doc-x',
          },
        ],
        totalQuantity: 3,
        totalAmount: 6,
        createdAt: '',
        updatedAt: '',
      },
    ]
    const rows = listInvoiceCandidatesFromOrders(orders, {})
    expect(rows).toHaveLength(1)
    expect(rows[0].materialName).toBe('要开')
  })
})
