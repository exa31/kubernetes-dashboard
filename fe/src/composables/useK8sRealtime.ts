import { ref } from 'vue'

import { useK8sStore } from '@/stores'
import { logger } from '@/utils'

export interface K8sChangeEvent {
  resource: 'secret' | 'configmap' | 'deployment' | 'service' | 'ingress' | 'cronjob' | 'pod'
  action: string
  namespace: string
  name: string
  timestamp: number
}

const isConnected = ref(false)
const lastEvent = ref<K8sChangeEvent | null>(null)
let eventSource: EventSource | null = null
let debounceTimer: number | null = null

export function useK8sRealtime() {
  const k8sStore = useK8sStore()

  function handleResourceUpdate(event: K8sChangeEvent) {
    lastEvent.value = event

    // Only auto-sync if event belongs to current namespace or cluster-wide
    if (event.namespace && event.namespace !== k8sStore.selectedNamespace) {
      return
    }

    if (debounceTimer !== null) {
      window.clearTimeout(debounceTimer)
    }

    debounceTimer = window.setTimeout(() => {
      logger.info('Auto-refreshing store due to real-time cluster event', event)
      switch (event.resource) {
        case 'secret':
          k8sStore.fetchSecrets()
          break
        case 'configmap':
          k8sStore.fetchConfigMaps()
          break
        case 'deployment':
        case 'pod':
          k8sStore.fetchDeployments()
          break
        case 'service':
          k8sStore.fetchServices()
          break
        case 'ingress':
          k8sStore.fetchIngresses()
          break
        case 'cronjob':
          k8sStore.fetchCronJobs()
          break
      }
    }, 400)
  }

  function connect() {
    if (typeof window === 'undefined' || eventSource) return

    // Connect to SSE events endpoint (proxied via Vite /api -> http://localhost:3001)
    const sseUrl = '/api/v1/realtime/sse/events'
    eventSource = new EventSource(sseUrl, { withCredentials: true })

    eventSource.onopen = () => {
      isConnected.value = true
      logger.info('Realtime K8s SSE stream connected')
    }

    eventSource.onmessage = (e) => {
      try {
        const parsed = JSON.parse(e.data)
        if (parsed.type === 'k8s_change' && parsed.data) {
          handleResourceUpdate(parsed.data as K8sChangeEvent)
        }
      } catch (err) {
        logger.debug('Non-JSON realtime event received', err)
      }
    }

    eventSource.onerror = () => {
      isConnected.value = false
    }
  }

  function disconnect() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
      isConnected.value = false
    }
  }

  return {
    isConnected,
    lastEvent,
    connect,
    disconnect,
  }
}
