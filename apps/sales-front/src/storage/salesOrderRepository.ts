import type { SalesOrder } from '@/types/salesOrder'
import { db } from './db'

export async function listAll(): Promise<SalesOrder[]> {
  return db.salesOrders.orderBy('updatedAt').reverse().toArray()
}

export async function getById(id: string): Promise<SalesOrder | undefined> {
  return db.salesOrders.get(id)
}

export async function put(order: SalesOrder): Promise<void> {
  await db.salesOrders.put(order)
}

export async function remove(id: string): Promise<void> {
  await db.salesOrders.delete(id)
}

export async function count(): Promise<number> {
  return db.salesOrders.count()
}

export async function putMany(orders: SalesOrder[]): Promise<void> {
  await db.salesOrders.bulkPut(orders)
}
