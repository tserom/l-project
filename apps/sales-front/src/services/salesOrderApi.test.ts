import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createSalesOrder,
  getSalesOrder,
  listSalesOrders,
  removeSalesOrder,
  updateSalesOrder,
} from './salesOrderApi'

const validInput = {
  orderNo: '00150262',
  orderDate: '2026-07-22',
  customerName: '884周村 马俊生',
  warehouseName: '中大慧科',
  deliveryType: '提货',
  lines: [
    {
      id: 'line-1',
      materialName: '002024-2Cr13黑棒',
      spec: 'Φ 32',
      unit: 'kg',
      quantity: 784,
      unitPrice: 6.9,
      amount: 0,
      needInvoice: false,
    },
  ],
}

function ok<T>(data: T) {
  return {
    ok: true,
    json: async () => ({ code: 0, message: 'ok', data }),
  }
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('salesOrderApi', () => {
  it('creates, lists, gets, updates, removes via HTTP', async () => {
    const created = {
      id: 'ord-1',
      ...validInput,
      totalQuantity: 784,
      totalAmount: 5409.6,
      createdAt: '2026-07-22T00:00:00Z',
      updatedAt: '2026-07-22T00:00:00Z',
      lines: [{ ...validInput.lines[0], amount: 5409.6 }],
    }

    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(ok(created))
      .mockResolvedValueOnce(
        ok({ list: [created], total: 1, page: 1, pageSize: 100 }),
      )
      .mockResolvedValueOnce(ok(created))
      .mockResolvedValueOnce(
        ok({ ...created, customerName: '新客户' }),
      )
      .mockResolvedValueOnce(ok({ ...created, customerName: '新客户' }))
      .mockResolvedValueOnce(ok({ ok: true }))
      .mockResolvedValueOnce(
        ok({ list: [], total: 0, page: 1, pageSize: 100 }),
      )

    vi.stubGlobal('fetch', fetchMock)

    const c = await createSalesOrder(validInput)
    expect(c.totalAmount).toBe(5409.6)
    expect((await listSalesOrders()).list).toHaveLength(1)
    expect((await getSalesOrder(c.id)).orderNo).toBe('00150262')
    await updateSalesOrder(c.id, { ...validInput, customerName: '新客户' })
    expect((await getSalesOrder(c.id)).customerName).toBe('新客户')
    await removeSalesOrder(c.id)
    expect((await listSalesOrders()).list).toHaveLength(0)
  })

  it('surfaces API error message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      json: async () => ({ code: 40000, message: '客户不能为空' }),
    }))
    await expect(
      createSalesOrder({ ...validInput, customerName: '  ' }),
    ).rejects.toThrow(/客户/)
  })
})
