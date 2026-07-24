import { describe, expect, it } from 'vitest'
import {
  calcLineAmount,
  formatAmount,
  formatQuantity,
  sumAmounts,
  sumQuantities,
} from './money'

describe('money', () => {
  it('calcLineAmount rounds to 2 decimals', () => {
    expect(calcLineAmount(784, 6.9)).toBe(5409.6)
  })

  it('sums lines', () => {
    const lines = [
      { quantity: 784, amount: 5409.6 },
      { quantity: 10, amount: 69 },
    ]
    expect(sumQuantities(lines)).toBe(794)
    expect(sumAmounts(lines)).toBe(5478.6)
  })

  it('formats like sample slip', () => {
    expect(formatQuantity(784)).toBe('784.00')
    expect(formatAmount(5409.6)).toBe('5,409.60')
  })
})
