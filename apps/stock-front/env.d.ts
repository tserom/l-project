/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare global {
  interface WujieAppProps {
    initialPath?: string
    instanceId?: string | null
    [key: string]: unknown
  }

  interface WujieBus {
    $on(event: string, handler: (raw?: unknown) => void): void
    $off(event: string, handler: (raw?: unknown) => void): void
    $emit(event: string, ...args: unknown[]): void
  }

  interface Window {
    __POWERED_BY_WUJIE__?: boolean
    __WUJIE_MOUNT?: () => void
    __WUJIE_UNMOUNT?: () => void
    $wujie?: {
      props?: WujieAppProps
      bus?: WujieBus | null
    }
  }
}

export {}
