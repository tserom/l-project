import { describe, expect, it } from 'vitest'
import { generateOrderNo } from './orderNo'

describe('generateOrderNo', () => {
  it('returns non-empty string of length >= 8', () => {
    const no = generateOrderNo(new Date('2026-07-22T12:00:00'))
    expect(no.length).toBeGreaterThanOrEqual(8)
    expect(no.startsWith('20260722')).toBe(true)
  })
})
