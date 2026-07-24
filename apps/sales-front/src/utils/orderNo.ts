export function generateOrderNo(now: Date = new Date()): string {
  const yyyy = String(now.getFullYear())
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  const suffix = String(now.getTime() % 10000).padStart(4, '0')
  return `${yyyy}${mm}${dd}${suffix}`
}
