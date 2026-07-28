import type { InvoiceDoc } from '@/types/invoiceDoc'
import { db } from './db'

export async function listAll(): Promise<InvoiceDoc[]> {
  return db.invoiceDocs.orderBy('updatedAt').reverse().toArray()
}

export async function getById(id: string): Promise<InvoiceDoc | undefined> {
  return db.invoiceDocs.get(id)
}

export async function put(doc: InvoiceDoc): Promise<void> {
  await db.invoiceDocs.put(doc)
}
