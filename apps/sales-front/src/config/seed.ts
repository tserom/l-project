import type { SalesOrder } from '@/types/salesOrder'

export const seedSalesOrders: SalesOrder[] = [
  {
    id: 'seed-1',
    orderNo: '00150262',
    orderDate: '2026-07-22',
    customerName: '884周村 马俊生',
    warehouseName: '01金阳仓库',
    deliveryType: '提货',
    lines: [
      {
        id: 'seed-line-1',
        materialName: '002024-2Cr13黑棒',
        spec: 'Φ 32',
        unit: 'kg',
        quantity: 784,
        unitPrice: 6.9,
        amount: 5409.6,
      },
    ],
    totalQuantity: 784,
    totalAmount: 5409.6,
    createdAt: '2026-07-22T00:00:00.000Z',
    updatedAt: '2026-07-22T00:00:00.000Z',
  },
]
