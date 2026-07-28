import type { InvoiceDocLine } from '@/types/invoiceDoc'
import { calcLineAmount } from '@/utils/money'

export type MergedInvoiceLine = {
  key: string
  materialName: string
  spec: string
  unit: string
  unitPrice: number
  quantity: number
  amount: number
  /** 合并了几条明细 */
  sourceCount: number
  orderNos: string
  orderDates: string
  customerNames: string
  warehouseNames: string
  deliveryTypes: string
}

function mergeKey(line: Pick<InvoiceDocLine, 'materialName' | 'spec' | 'unitPrice'>): string {
  return `${line.materialName}\0${line.spec}\0${line.unitPrice}`
}

function uniqueJoin(values: string[]): string {
  return [...new Set(values.filter(Boolean))].join('、')
}

/** 按物资 + 规格型号 + 单价 合并；数量相加，金额按合计数量×单价重算 */
export function mergeInvoiceLines(lines: InvoiceDocLine[]): MergedInvoiceLine[] {
  const map = new Map<
    string,
    {
      materialName: string
      spec: string
      unit: string
      unitPrice: number
      quantity: number
      orderNos: string[]
      orderDates: string[]
      customerNames: string[]
      warehouseNames: string[]
      deliveryTypes: string[]
      sourceCount: number
    }
  >()

  for (const line of lines) {
    const key = mergeKey(line)
    const existing = map.get(key)
    if (!existing) {
      map.set(key, {
        materialName: line.materialName,
        spec: line.spec,
        unit: line.unit,
        unitPrice: line.unitPrice,
        quantity: line.quantity,
        orderNos: [line.orderNo],
        orderDates: [line.orderDate],
        customerNames: [line.customerName],
        warehouseNames: [line.warehouseName],
        deliveryTypes: [line.deliveryType],
        sourceCount: 1,
      })
    } else {
      existing.quantity += line.quantity
      existing.sourceCount += 1
      existing.orderNos.push(line.orderNo)
      existing.orderDates.push(line.orderDate)
      existing.customerNames.push(line.customerName)
      existing.warehouseNames.push(line.warehouseName)
      existing.deliveryTypes.push(line.deliveryType)
    }
  }

  return [...map.entries()].map(([key, row]) => ({
    key,
    materialName: row.materialName,
    spec: row.spec,
    unit: row.unit,
    unitPrice: row.unitPrice,
    quantity: row.quantity,
    amount: calcLineAmount(row.quantity, row.unitPrice),
    sourceCount: row.sourceCount,
    orderNos: uniqueJoin(row.orderNos),
    orderDates: uniqueJoin(row.orderDates),
    customerNames: uniqueJoin(row.customerNames),
    warehouseNames: uniqueJoin(row.warehouseNames),
    deliveryTypes: uniqueJoin(row.deliveryTypes),
  }))
}
