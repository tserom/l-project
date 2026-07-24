/** ISO `YYYY-MM-DD` → `YYYY年M月D日`（月日不补零） */
export function formatOrderDate(isoDate: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(isoDate.trim())
  if (!m) {
    throw new Error(`无效日期: ${isoDate}`)
  }
  const year = m[1]
  const month = Number(m[2])
  const day = Number(m[3])
  return `${year}年${month}月${day}日`
}
