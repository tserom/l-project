<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  VUE_PROJECT_ROUTE_EVENT,
  SUB_ROUTE_CHANGE_EVENT,
  SUB_APP_NAME,
  getWujieProps,
  getWujieBus,
  parseRoutePayload,
  normalizePath,
  whenWujieBusReady,
} from '@/utils/wujie'

const router = useRouter()
const route = useRoute()
const skipSyncToParent = ref(true)

let offRoute: (() => void) | undefined
let stopWaiting: (() => void) | undefined
const propTimers: ReturnType<typeof setTimeout>[] = []

function applyFromProps() {
  const raw = getWujieProps().initialPath
  if (typeof raw !== 'string' || raw === '/') return
  const target = normalizePath(raw)
  if (window.location.pathname !== target) {
    router.replace(target)
  }
}

onMounted(() => {
  applyFromProps()
  ;[0, 50, 200].forEach((ms) => {
    propTimers.push(setTimeout(applyFromProps, ms))
  })

  stopWaiting = whenWujieBusReady((bus) => {
    const onRoute = (raw: unknown) => {
      const { path, instanceId } = parseRoutePayload(raw)
      if (typeof path !== 'string') return
      const myId = getWujieProps().instanceId
      if (instanceId != null && myId !== instanceId) return
      const target = normalizePath(path)
      if (window.location.pathname !== target) {
        router.push(target)
      }
    }

    bus.$on(VUE_PROJECT_ROUTE_EVENT, onRoute)
    offRoute = () => bus.$off(VUE_PROJECT_ROUTE_EVENT, onRoute)

    const initPath = getWujieProps().initialPath
    if (typeof initPath === 'string' && initPath !== '/') {
      onRoute({ path: initPath, instanceId: getWujieProps().instanceId })
    }
  })
})

onUnmounted(() => {
  propTimers.forEach(clearTimeout)
  stopWaiting?.()
  offRoute?.()
})

watch(
  () => route.path,
  (pathname) => {
    const bus = getWujieBus()
    if (!bus) return
    if (skipSyncToParent.value) {
      skipSyncToParent.value = false
      return
    }
    bus.$emit(SUB_ROUTE_CHANGE_EVENT, SUB_APP_NAME, {
      path: pathname,
      instanceId: getWujieProps().instanceId,
    })
  },
)
</script>

<template></template>
