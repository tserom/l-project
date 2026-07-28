import Dexie, { type EntityTable } from 'dexie'
import type { InvoiceDoc } from '@/types/invoiceDoc'
import type { SalesOrder } from '@/types/salesOrder'

export class SalesFrontDB extends Dexie {
  salesOrders!: EntityTable<SalesOrder, 'id'>
  invoiceDocs!: EntityTable<InvoiceDoc, 'id'>

  constructor() {
    super('sales-front')
    this.version(1).stores({
      salesOrders: 'id, orderNo, updatedAt',
    })
    this.version(2).stores({
      salesOrders: 'id, orderNo, updatedAt',
      invoiceDocs: 'id, invoiceNo, status, updatedAt',
    })
  }
}

export const db = new SalesFrontDB()
