import { describe, expect, it, beforeEach } from 'vitest'
import {
  createInvoiceDoc,
  listInvoiceCandidates,
  listInvoiceCandidatesFromOrders,
  voidInvoiceDoc,
} from './invoiceDocApi'
import { createSalesOrder } from './salesOrderApi'
import { db } from '@/storage/db'
import type { SalesOrder } from '@/types/salesOrder'

beforeEach(async () => {
  await db.delete()
  await db.open()
})

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

describe('createInvoiceDoc / voidInvoiceDoc', () => {
  it('marks lines then clears on void', async () => {
    const order = await createSalesOrder({
      orderNo: 'T001',
      orderDate: '2026-07-10',
      customerName: '客户乙',
      warehouseName: '01金阳仓库',
      deliveryType: '提货',
      lines: [
        {
          id: 'line-a',
          materialName: '棒料',
          spec: 'Φ10',
          unit: 'kg',
          quantity: 10,
          unitPrice: 5,
          amount: 50,
          needInvoice: true,
        },
      ],
    })

    const candidates = await listInvoiceCandidates({ customerName: '客户乙' })
    expect(candidates).toHaveLength(1)

    const doc = await createInvoiceDoc({
      filterCustomerName: '客户乙',
      lines: candidates,
    })
    expect(doc.status).toBe('saved')
    expect(doc.totalAmount).toBe(50)

    const after = await listInvoiceCandidates({ customerName: '客户乙' })
    expect(after).toHaveLength(0)

    await voidInvoiceDoc(doc.id)
    const again = await listInvoiceCandidates({ customerName: '客户乙' })
    expect(again).toHaveLength(1)
    expect(again[0].salesOrderLineId).toBe(order.lines[0].id)
  })
})
