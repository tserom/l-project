export interface SalesOrderLine {
  id: string
  materialName: string
  spec: string
  unit: string
  quantity: number
  unitPrice: number
  amount: number
  lineRemark?: string
}

export interface SalesOrder {
  id: string
  orderNo: string
  orderDate: string
  customerName: string
  warehouseName: string
  deliveryType: string
  remark?: string
  lines: SalesOrderLine[]
  totalQuantity: number
  totalAmount: number
  createdAt: string
  updatedAt: string
}

/** 创建/更新时由调用方提供的字段（合计与时间戳由 Api 写入） */
export type SalesOrderInput = Omit<
  SalesOrder,
  'id' | 'totalQuantity' | 'totalAmount' | 'createdAt' | 'updatedAt'
> & { id?: string }
