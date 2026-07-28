export type InvoiceDocStatus = 'saved' | 'voided'

export interface InvoiceDocLine {
  id: string
  salesOrderId: string
  salesOrderLineId: string
  materialName: string
  spec: string
  unit: string
  quantity: number
  unitPrice: number
  amount: number
  lineRemark?: string
  /** 行末展示：来源销售单快照 */
  orderNo: string
  orderDate: string
  customerName: string
  warehouseName: string
  deliveryType: string
}

export interface InvoiceDoc {
  id: string
  invoiceNo: string
  status: InvoiceDocStatus
  filterCustomerName?: string
  filterDateFrom?: string
  filterDateTo?: string
  filterOrderNos?: string[]
  lines: InvoiceDocLine[]
  totalQuantity: number
  totalAmount: number
  createdAt: string
  updatedAt: string
  voidedAt?: string
}

export type InvoiceCandidateFilter = {
  customerName?: string
  dateFrom?: string
  dateTo?: string
  orderNos?: string[]
}

/** 加载候选时的展平行（尚未写入开票单） */
export type InvoiceCandidateLine = InvoiceDocLine
