/**
 * 入口索引
 * | 动作 | 函数 | 存储 |
 * | 列表 | listSalesOrders | IndexedDB salesOrders |
 * | 详情 | getSalesOrder | 同上 |
 * | 新建 | createSalesOrder | 同上 |
 * | 更新 | updateSalesOrder | 同上 |
 * | 删除 | removeSalesOrder | 同上 |
 * | 种子 | ensureSeedData | 仅空库 |
 */
import type { SalesOrder, SalesOrderInput, SalesOrderLine } from '@/types/salesOrder'
import {
  calcLineAmount,
  sumAmounts,
  sumQuantities,
} from '@/utils/money'
import * as repo from '@/storage/salesOrderRepository'

function normalizeLines(lines: SalesOrderLine[]): SalesOrderLine[] {
  return lines.map((line) => ({
    ...line,
    amount: calcLineAmount(line.quantity, line.unitPrice),
  }))
}

function validateInput(input: SalesOrderInput): void {
  if (!input.customerName.trim()) {
    throw new Error('客户不能为空')
  }
  if (!input.lines.length) {
    throw new Error('至少需要一行明细')
  }
  for (const line of input.lines) {
    if (line.quantity < 0) {
      throw new Error('数量不能为负数')
    }
    if (line.unitPrice < 0) {
      throw new Error('单价不能为负数')
    }
  }
}

function buildOrder(
  id: string,
  input: SalesOrderInput,
  createdAt: string,
  updatedAt: string,
): SalesOrder {
  const lines = normalizeLines(input.lines)
  return {
    id,
    orderNo: input.orderNo,
    orderDate: input.orderDate,
    customerName: input.customerName.trim(),
    warehouseName: input.warehouseName,
    deliveryType: input.deliveryType,
    remark: input.remark,
    lines,
    totalQuantity: sumQuantities(lines),
    totalAmount: sumAmounts(lines),
    createdAt,
    updatedAt,
  }
}

export async function listSalesOrders(): Promise<SalesOrder[]> {
  return repo.listAll()
}

export async function getSalesOrder(id: string): Promise<SalesOrder> {
  const order = await repo.getById(id)
  if (!order) {
    throw new Error('销售单不存在')
  }
  return order
}

export async function createSalesOrder(input: SalesOrderInput): Promise<SalesOrder> {
  validateInput(input)
  const now = new Date().toISOString()
  const order = buildOrder(crypto.randomUUID(), input, now, now)
  await repo.put(order)
  return order
}

export async function updateSalesOrder(
  id: string,
  input: SalesOrderInput,
): Promise<SalesOrder> {
  validateInput(input)
  const existing = await repo.getById(id)
  if (!existing) {
    throw new Error('销售单不存在')
  }
  const order = buildOrder(id, input, existing.createdAt, new Date().toISOString())
  await repo.put(order)
  return order
}

export async function removeSalesOrder(id: string): Promise<void> {
  await repo.remove(id)
}

export async function ensureSeedData(seeds: SalesOrder[]): Promise<void> {
  if ((await repo.count()) > 0) {
    return
  }
  if (seeds.length === 0) {
    return
  }
  await repo.putMany(seeds)
}
