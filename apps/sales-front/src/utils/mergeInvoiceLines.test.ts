import { describe, expect, it } from 'vitest'
import type { InvoiceDocLine } from '@/types/invoiceDoc'
import { mergeInvoiceLines } from './mergeInvoiceLines'

function line(partial: Partial<InvoiceDocLine> & Pick<InvoiceDocLine, 'id' | 'materialName' | 'spec' | 'unitPrice' | 'quantity' | 'amount'>): InvoiceDocLine {
  return {
    salesOrderId: 'o1',
    salesOrderLineId: partial.id,
    unit: 'kg',
    orderNo: 'A',
    orderDate: '2026-07-01',
    customerName: '甲',
    warehouseName: '仓',
    deliveryType: '提货',
    ...partial,
  }
}

describe('mergeInvoiceLines', () => {
  it('merges same material+unitPrice and joins different specs', () => {
    const merged = mergeInvoiceLines([
      line({
        id: '1',
        materialName: '圆钢',
        spec: '50',
        unitPrice: 10,
        quantity: 1,
        amount: 10,
      }),
      line({
        id: '2',
        materialName: '圆钢',
        spec: '40',
        unitPrice: 10,
        quantity: 2,
        amount: 20,
      }),
      line({
        id: '3',
        materialName: '圆钢',
        spec: '30',
        unitPrice: 10,
        quantity: 3,
        amount: 30,
      }),
      line({
        id: '4',
        materialName: '圆钢',
        spec: '30',
        unitPrice: 11,
        quantity: 1,
        amount: 11,
      }),
    ])
    expect(merged).toHaveLength(2)
    const at10 = merged.find((r) => r.unitPrice === 10)!
    expect(at10.spec).toBe('30-40-50')
    expect(at10.quantity).toBe(6)
    expect(at10.amount).toBe(60)
    expect(at10.sourceCount).toBe(3)
  })

  it('merges same material+spec+unitPrice and sums qty', () => {
    const merged = mergeInvoiceLines([
      line({
        id: '1',
        materialName: '棒',
        spec: 'Φ32',
        unitPrice: 6.9,
        quantity: 100,
        amount: 690,
        orderNo: 'A1',
      }),
      line({
        id: '2',
        materialName: '棒',
        spec: 'Φ32',
        unitPrice: 6.9,
        quantity: 50,
        amount: 345,
        orderNo: 'A2',
        orderDate: '2026-07-02',
      }),
      line({
        id: '3',
        materialName: '棒',
        spec: 'Φ32',
        unitPrice: 7,
        quantity: 10,
        amount: 70,
      }),
    ])
    expect(merged).toHaveLength(2)
    const first = merged.find((r) => r.unitPrice === 6.9)!
    expect(first.quantity).toBe(150)
    expect(first.amount).toBe(1035)
    expect(first.sourceCount).toBe(2)
    expect(first.orderNos).toContain('A1')
    expect(first.orderNos).toContain('A2')
  })
})
