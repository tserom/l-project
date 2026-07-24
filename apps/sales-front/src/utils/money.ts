function round2(n: number): number {
  return Math.round(n * 100) / 100
}

export function calcLineAmount(quantity: number, unitPrice: number): number {
  return round2(quantity * unitPrice)
}

export function sumQuantities(lines: Array<{ quantity: number }>): number {
  return round2(lines.reduce((sum, line) => sum + line.quantity, 0))
}

export function sumAmounts(lines: Array<{ amount: number }>): number {
  return round2(lines.reduce((sum, line) => sum + line.amount, 0))
}

export function formatQuantity(n: number): string {
  return n.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

export function formatAmount(n: number): string {
  return n.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}
