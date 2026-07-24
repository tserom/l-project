import type { SalesOrderLine } from '@/types/salesOrder'

export function padLinesToMin(
  lines: SalesOrderLine[],
  minRows: number,
): Array<SalesOrderLine | null> {
  const result: Array<SalesOrderLine | null> = [...lines]
  while (result.length < minRows) {
    result.push(null)
  }
  return result
}
