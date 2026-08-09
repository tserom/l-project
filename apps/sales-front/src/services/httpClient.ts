type Envelope<T> = {
  code: number
  message: string
  data?: T
}

/**
 * Call sales-manage API (Vite proxies `/api` → :8083 in dev).
 */
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  if (init?.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(path, { ...init, headers })
  let body: Envelope<T>
  try {
    body = (await res.json()) as Envelope<T>
  } catch {
    throw new Error(`请求失败 (${res.status})`)
  }
  if (!res.ok || body.code !== 0) {
    throw new Error(body.message || `请求失败 (${res.status})`)
  }
  return body.data as T
}

export type PageResult<T> = {
  list: T[]
  total: number
  page: number
  pageSize: number
}

/** Append qp-* / page params to a path. */
export function withQuery(
  path: string,
  params: Record<string, string | number | undefined | null>,
): string {
  const sp = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    sp.set(key, String(value))
  }
  const q = sp.toString()
  return q ? `${path}?${q}` : path
}
