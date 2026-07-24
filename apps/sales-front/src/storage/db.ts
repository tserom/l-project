import Dexie, { type EntityTable } from 'dexie'
import type { SalesOrder } from '@/types/salesOrder'

export class SalesFrontDB extends Dexie {
  salesOrders!: EntityTable<SalesOrder, 'id'>

  constructor() {
    super('sales-front')
    this.version(1).stores({
      salesOrders: 'id, orderNo, updatedAt',
    })
  }
}

export const db = new SalesFrontDB()
