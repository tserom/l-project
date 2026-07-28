import { beforeEach, describe, expect, it } from 'vitest'
import { seedSalesOrders } from '@/config/seed'
import { db } from '@/storage/db'
import {
  createSalesOrder,
  ensureSeedData,
  getSalesOrder,
  listSalesOrders,
  removeSalesOrder,
  updateSalesOrder,
} from './salesOrderApi'

const validInput = {
  orderNo: '00150262',
  orderDate: '2026-07-22',
  customerName: '884周村 马俊生',
  warehouseName: '01金阳仓库',
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

beforeEach(async () => {
  await db.delete()
  await db.open()
})

describe('salesOrderApi', () => {
  it('creates, lists, gets, updates, removes', async () => {
    const created = await createSalesOrder(validInput)
    expect(created.totalAmount).toBe(5409.6)
    expect(created.totalQuantity).toBe(784)
    expect(await listSalesOrders()).toHaveLength(1)
    const got = await getSalesOrder(created.id)
    expect(got.orderNo).toBe('00150262')
    await updateSalesOrder(created.id, { ...validInput, customerName: '新客户' })
    expect((await getSalesOrder(created.id)).customerName).toBe('新客户')
    await removeSalesOrder(created.id)
    expect(await listSalesOrders()).toHaveLength(0)
  })

  it('rejects empty customer', async () => {
    await expect(
      createSalesOrder({ ...validInput, customerName: '  ' }),
    ).rejects.toThrow(/客户/)
  })

  it('ensureSeedData only when empty', async () => {
    await ensureSeedData(seedSalesOrders)
    await ensureSeedData(seedSalesOrders)
    expect(await listSalesOrders()).toHaveLength(seedSalesOrders.length)
  })
})
