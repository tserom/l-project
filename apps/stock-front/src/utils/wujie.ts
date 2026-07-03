/**
 * wujie 微前端：与 Host、导航 subAppBusName 保持一致
 */
export const IS_WUJIE =
  typeof window !== 'undefined' && !!window.__POWERED_BY_WUJIE__

export const VUE_PROJECT_ROUTE_EVENT = 'vue-project-route'
export const SUB_ROUTE_CHANGE_EVENT = 'sub-route-change'
export const SUB_APP_NAME = 'vue-project'

export const getWujieProps = (): WujieAppProps =>
  (typeof window !== 'undefined' && window.$wujie?.props) || {}

export const getWujieBus = (): WujieBus | null =>
  typeof window !== 'undefined' ? (window.$wujie?.bus ?? null) : null

export function parseRoutePayload(raw: unknown): {
  path: string | null
  instanceId: string | null
} {
  if (typeof raw === 'string') return { path: raw, instanceId: null }
  if (raw && typeof raw === 'object') {
    const o = raw as { path?: unknown; instanceId?: unknown }
    const idVal = o.instanceId
    return {
      path: typeof o.path === 'string' ? o.path : null,
      instanceId:
        idVal == null
          ? null
          : typeof idVal === 'string'
            ? idVal
            : String(idVal),
    }
  }
  return { path: null, instanceId: null }
}

export const normalizePath = (path: string | undefined | null): string =>
  path?.startsWith('/') ? path : `/${path ?? ''}`

export function whenWujieBusReady(
  onReady: (bus: WujieBus) => void,
  { interval = 15, maxAttempts = 200 }: { interval?: number; maxAttempts?: number } = {},
): () => void {
  let cancelled = false
  let attempts = 0

  const tick = () => {
    if (cancelled) return
    const bus = getWujieBus()
    if (bus) {
      onReady(bus)
      return
    }
    if (attempts++ < maxAttempts) {
      setTimeout(tick, interval)
    }
  }
  tick()

  return () => {
    cancelled = true
  }
}
