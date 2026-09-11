import { ref } from 'vue'

import { useK8sStore } from '@/stores'
import { logger } from '@/utils'

export interface K8sChangeEvent {
  resource:
    | 'secret'
    | 'configmap'
    | 'deployment'
    | 'statefulset'
    | 'daemonset'
    | 'service'
    | 'ingress'
    | 'cronjob'
    | 'pod'
    | 'pvc'
    | 'pv'
    | 'namespace'
  action: string
  namespace: string
  name: string
  timestamp: number
}

const isConnected = ref(false)
const lastEvent = ref<K8sChangeEvent | null>(null)
const lastSyncTime = ref<Date>(new Date())

let eventSource: EventSource | null = null
let debounceTimer: number | null = null
let reconnectAttempts = 0
let reconnectTimeout: number | null = null

export function useK8sRealtime() {
  const k8sStore = useK8sStore()

  function handleResourceUpdate(event: K8sChangeEvent) {
    lastEvent.value = event
    lastSyncTime.value = new Date()

    // Only auto-sync if event belongs to current namespace or is cluster-wide
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
          k8sStore.fetchDeployments()
          k8sStore.fetchPods()
          break
        case 'statefulset':
          k8sStore.fetchStatefulSets()
          k8sStore.fetchPods()
          break
        case 'daemonset':
          k8sStore.fetchDaemonSets()
          k8sStore.fetchPods()
          break
        case 'pod':
          k8sStore.fetchPods()
          k8sStore.fetchPodMetrics()
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
        case 'pvc':
        case 'pv':
          k8sStore.fetchPVCs()
          k8sStore.fetchPVs()
          break
        case 'namespace':
          k8sStore.fetchNamespaces()
          k8sStore.fetchClusterOverview()
          break
      }
    }, 300)
  }

  function scheduleReconnect() {
    if (reconnectTimeout !== null || typeof window === 'undefined') return
    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 15000)
    reconnectAttempts++
    logger.debug(`Scheduling SSE reconnect in ${delay}ms (attempt ${reconnectAttempts})`)
    reconnectTimeout = window.setTimeout(() => {
      reconnectTimeout = null
      connect()
    }, delay)
  }

  function connect() {
    if (typeof window === 'undefined') return
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }

    // Connect to SSE events endpoint (proxied via Vite /api -> http://localhost:3001)
    const sseUrl = '/api/v1/realtime/sse/events'
    try {
      eventSource = new EventSource(sseUrl, { withCredentials: true })

      eventSource.onopen = () => {
        isConnected.value = true
        reconnectAttempts = 0
        if (reconnectTimeout !== null) {
          window.clearTimeout(reconnectTimeout)
          reconnectTimeout = null
        }
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
        if (eventSource) {
          eventSource.close()
          eventSource = null
        }
        scheduleReconnect()
      }
    } catch {
      isConnected.value = false
      scheduleReconnect()
    }
  }

  function disconnect() {
    if (reconnectTimeout !== null) {
      window.clearTimeout(reconnectTimeout)
      reconnectTimeout = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    isConnected.value = false
  }

  return {
    isConnected,
    lastEvent,
    lastSyncTime,
    connect,
    disconnect
  }
}
