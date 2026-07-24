import { describe, expect, it } from 'vitest'
import { formatOrderDate } from './dateFormat'

describe('formatOrderDate', () => {
  it('formats ISO date as Chinese', () => {
    expect(formatOrderDate('2026-07-22')).toBe('2026年7月22日')
  })
})
