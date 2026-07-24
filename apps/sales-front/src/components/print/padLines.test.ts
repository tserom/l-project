import { describe, expect, it } from 'vitest'
import type { SalesOrderLine } from '@/types/salesOrder'
import { padLinesToMin } from './padLines'

describe('padLinesToMin', () => {
  it('pads with null to min rows', () => {
    expect(padLinesToMin([], 3)).toEqual([null, null, null])
    expect(padLinesToMin([{ id: '1' } as SalesOrderLine], 2)).toHaveLength(2)
  })
})
