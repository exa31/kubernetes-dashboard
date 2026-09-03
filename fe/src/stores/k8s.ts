import { defineStore } from 'pinia'
import { ref } from 'vue'

import { k8sApi } from '@/api'
import type {
  ClusterInfo,
  ClusterOverview,
  ConfigMapDetail,
  ConfigMapItem,
  CreateCronJobPayload,
  CreateNamespacePayload,
  CronJobDetail,
  CronJobItem,
  DaemonSetItem,
  DeploymentDetail,
  DeploymentItem,
  EventItem,
  IngressItem,
  JobItem,
  NamespaceItem,
  NodeItem,
  PodItem,
  PodLogsResponse,
  PodMetrics,
  PVCItem,
  PVItem,
  ResourceQuotaItem,
  ResourceYAMLResponse,
  RolloutRestartResponse,
  SaveConfigMapPayload,
  SaveSecretPayload,
  SecretDetail,
  SecretItem,
  ServiceDetail,
  ServiceItem,
  StatefulSetItem,
  UpdateCronJobPayload,
  UpdateDeploymentPayload,
} from '@/types'
import { logger } from '@/utils'

export const useK8sStore = defineStore('k8s', () => {
  const clusterInfo = ref<ClusterInfo | null>(null)
  const namespaces = ref<NamespaceItem[]>([])
  const selectedNamespace = ref<string>(localStorage.getItem('kubeenv_selected_ns') || 'dev-coffe')

  const secrets = ref<SecretItem[]>([])
  const configmaps = ref<ConfigMapItem[]>([])
  const deployments = ref<DeploymentItem[]>([])
  const statefulsets = ref<StatefulSetItem[]>([])
  const daemonsets = ref<DaemonSetItem[]>([])
  const pods = ref<PodItem[]>([])
  const services = ref<ServiceItem[]>([])
  const ingresses = ref<IngressItem[]>([])
  const cronjobs = ref<CronJobItem[]>([])
  const events = ref<EventItem[]>([])
  const pvcs = ref<PVCItem[]>([])
  const pvs = ref<PVItem[]>([])
  const clusterOverview = ref<ClusterOverview | null>(null)
  const nodes = ref<NodeItem[]>([])
  const podMetrics = ref<Record<string, PodMetrics>>({})
  const resourceQuotas = ref<ResourceQuotaItem[]>([])
  const eventsFeed = ref<EventItem[]>([])

  const isLoading = ref(false)
  const isActionLoading = ref(false)
  const error = ref<string | null>(null)

  function setNamespace(ns: string) {
    selectedNamespace.value = ns
    localStorage.setItem('kubeenv_selected_ns', ns)
    fetchAllResources(ns)
  }

  function getErrorMessage(err: unknown, fallback: string): string {
    if (err && typeof err === 'object' && 'response' in err) {
      const res = (err as { response?: { data?: { message?: string } } }).response
      if (res?.data?.message) return res.data.message
    }
    if (err instanceof Error) return err.message
    return fallback
  }

  async function fetchClusterInfo() {
    try {
      clusterInfo.value = await k8sApi.getClusterInfo()
    } catch (err: unknown) {
      logger.warn('Failed to fetch cluster info', err)
      error.value = getErrorMessage(err, 'Failed to connect to cluster')
    }
  }

  async function fetchNamespaces() {
    try {
      const list = await k8sApi.getNamespaces()
      namespaces.value = list
      // If current selected namespace not in list, select first available or default
      if (list.length > 0 && !list.some((n) => n.name === selectedNamespace.value)) {
        const preferred = list.find((n) => n.name === 'dev-coffe') || list[0]
        selectedNamespace.value = preferred.name
        localStorage.setItem('kubeenv_selected_ns', preferred.name)
      }
    } catch (err: unknown) {
      logger.warn('Failed to fetch namespaces', err)
    }
  }

  async function fetchSecrets(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      secrets.value = await k8sApi.getSecrets(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch secrets', err)
      error.value = getErrorMessage(err, 'Failed to fetch secrets')
    } finally {
      isLoading.value = false
    }
  }

  async function fetchConfigMaps(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      configmaps.value = await k8sApi.getConfigMaps(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch configmaps', err)
      error.value = getErrorMessage(err, 'Failed to fetch configmaps')
    } finally {
      isLoading.value = false
    }
  }

  async function fetchDeployments(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      deployments.value = await k8sApi.getDeployments(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch deployments', err)
      error.value = getErrorMessage(err, 'Failed to fetch deployments')
    } finally {
      isLoading.value = false
    }
  }

  async function fetchServices(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      services.value = await k8sApi.getServices(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch services', err)
      error.value = getErrorMessage(err, 'Failed to fetch services')
    } finally {
      isLoading.value = false
    }
  }

  async function fetchIngresses(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      ingresses.value = await k8sApi.getIngresses(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch ingresses', err)
      error.value = getErrorMessage(err, 'Failed to fetch ingresses')
    } finally {
      isLoading.value = false
    }
  }

  async function fetchCronJobs(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      cronjobs.value = await k8sApi.getCronJobs(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch cronjobs', err)
      error.value = getErrorMessage(err, 'Failed to fetch cronjobs')
    } finally {
      isLoading.value = false
    }
  }

  async function fetchAllResources(ns: string = selectedNamespace.value) {
    isLoading.value = true
    error.value = null
    try {
      await Promise.allSettled([
        fetchSecrets(ns),
        fetchConfigMaps(ns),
        fetchDeployments(ns),
        fetchStatefulSets(ns),
        fetchDaemonSets(ns),
        fetchPods(ns),
        fetchServices(ns),
        fetchIngresses(ns),
        fetchCronJobs(ns),
        fetchClusterInfo(),
        fetchClusterOverview(),
      ])
    } finally {
      isLoading.value = false
    }
  }

  async function getCronJobDetail(name: string, ns: string = selectedNamespace.value): Promise<CronJobDetail> {
    return await k8sApi.getCronJob(ns, name)
  }

  async function createCronJob(payload: CreateCronJobPayload): Promise<CronJobDetail> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.createCronJob(payload)
      await fetchCronJobs(payload.namespace)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function updateCronJob(name: string, payload: UpdateCronJobPayload, ns: string = selectedNamespace.value): Promise<CronJobDetail> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.updateCronJob(ns, name, payload)
      await fetchCronJobs(ns)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function toggleSuspendCronJob(name: string, ns: string = selectedNamespace.value): Promise<boolean> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.toggleSuspendCronJob(ns, name)
      await fetchCronJobs(ns)
      return res.suspended
    } finally {
      isActionLoading.value = false
    }
  }

  async function triggerCronJobNow(name: string, ns: string = selectedNamespace.value): Promise<JobItem> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.triggerCronJobNow(ns, name)
      await fetchCronJobs(ns)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function getCronJobJobs(name: string, ns: string = selectedNamespace.value): Promise<JobItem[]> {
    return await k8sApi.getCronJobJobs(ns, name)
  }

  async function deleteCronJob(name: string, ns: string = selectedNamespace.value): Promise<void> {
    isActionLoading.value = true
    try {
      await k8sApi.deleteCronJob(ns, name)
      await fetchCronJobs(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function getServiceDetail(name: string, ns: string = selectedNamespace.value): Promise<ServiceDetail> {
    return await k8sApi.getService(ns, name)
  }

  async function getIngressDetail(name: string, ns: string = selectedNamespace.value): Promise<IngressItem> {
    return await k8sApi.getIngress(ns, name)
  }

  async function fetchEvents(ns: string = selectedNamespace.value) {
    try {
      events.value = await k8sApi.getEvents(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch events', err)
    }
  }

  async function fetchPVCs(ns: string = selectedNamespace.value) {
    isLoading.value = true
    try {
      pvcs.value = await k8sApi.getPVCs(ns)
    } catch (err: unknown) {
      logger.error('Failed to fetch PVCs', err)
    } finally {
      isLoading.value = false
    }
  }

  async function fetchPVs() {
    isLoading.value = true
    try {
      pvs.value = await k8sApi.getPVs()
    } catch (err: unknown) {
      logger.error('Failed to fetch PVs', err)
    } finally {
      isLoading.value = false
    }
  }

  async function scaleDeployment(name: string, replicas: number, ns: string = selectedNamespace.value): Promise<DeploymentDetail> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.scaleDeployment(ns, name, replicas)
      await fetchDeployments(ns)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function getSecretDetail(name: string, ns: string = selectedNamespace.value): Promise<SecretDetail> {
    return await k8sApi.getSecret(ns, name)
  }

  async function saveSecret(payload: SaveSecretPayload): Promise<SecretDetail> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.saveSecret(payload)
      await fetchSecrets(payload.namespace)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function deleteSecret(name: string, ns: string = selectedNamespace.value) {
    isActionLoading.value = true
    try {
      await k8sApi.deleteSecret(ns, name)
      await fetchSecrets(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function getConfigMapDetail(name: string, ns: string = selectedNamespace.value): Promise<ConfigMapDetail> {
    return await k8sApi.getConfigMap(ns, name)
  }

  async function saveConfigMap(payload: SaveConfigMapPayload): Promise<ConfigMapDetail> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.saveConfigMap(payload)
      await fetchConfigMaps(payload.namespace)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function deleteConfigMap(name: string, ns: string = selectedNamespace.value) {
    isActionLoading.value = true
    try {
      await k8sApi.deleteConfigMap(ns, name)
      await fetchConfigMaps(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function restartDeployment(name: string, ns: string = selectedNamespace.value): Promise<RolloutRestartResponse> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.restartDeployment(ns, name)
      await fetchDeployments(ns)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function getDeploymentDetail(name: string, ns: string = selectedNamespace.value): Promise<DeploymentDetail> {
    return await k8sApi.getDeploymentDetail(ns, name)
  }

  async function updateDeployment(
    name: string,
    payload: UpdateDeploymentPayload,
    ns: string = selectedNamespace.value,
  ): Promise<DeploymentDetail> {
    isActionLoading.value = true
    try {
      const res = await k8sApi.updateDeployment(ns, name, payload)
      await fetchDeployments(ns)
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function getDeploymentPods(name: string, ns: string = selectedNamespace.value): Promise<PodItem[]> {
    return await k8sApi.getDeploymentPods(ns, name)
  }

  async function getPodLogs(
    podName: string,
    params?: { container?: string; tail_lines?: number; timestamps?: boolean },
    ns: string = selectedNamespace.value,
  ): Promise<PodLogsResponse> {
    return await k8sApi.getPodLogs(ns, podName, params)
  }

  async function applyYAML(yamlContent: string, targetNs: string = selectedNamespace.value, dryRun = false) {
    isActionLoading.value = true
    try {
      const res = await k8sApi.applyYAML(yamlContent, targetNs, dryRun)
      if (!dryRun && res.success_count > 0) {
        // Optimistically refresh dashboard resources
        fetchAllResources(targetNs)
      }
      return res
    } finally {
      isActionLoading.value = false
    }
  }

  async function fetchClusterOverview() {
    try {
      clusterOverview.value = await k8sApi.getClusterOverview()
      if (clusterOverview.value?.nodes) {
        nodes.value = clusterOverview.value.nodes
      }
    } catch (err) {
      logger.error('Failed to fetch cluster overview', err)
    }
  }

  async function fetchNodes() {
    try {
      nodes.value = await k8sApi.getNodes()
    } catch (err) {
      logger.error('Failed to fetch nodes', err)
    }
  }

  async function fetchStatefulSets(ns: string = selectedNamespace.value) {
    try {
      statefulsets.value = await k8sApi.getStatefulSets(ns)
    } catch (err) {
      logger.error('Failed to fetch statefulsets', err)
    }
  }

  async function scaleStatefulSet(name: string, replicas: number, ns: string = selectedNamespace.value) {
    isActionLoading.value = true
    try {
      await k8sApi.scaleStatefulSet(ns, name, replicas)
      await fetchStatefulSets(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function restartStatefulSet(name: string, ns: string = selectedNamespace.value) {
    isActionLoading.value = true
    try {
      await k8sApi.restartStatefulSet(ns, name)
      await fetchStatefulSets(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function fetchDaemonSets(ns: string = selectedNamespace.value) {
    try {
      daemonsets.value = await k8sApi.getDaemonSets(ns)
    } catch (err) {
      logger.error('Failed to fetch daemonsets', err)
    }
  }

  async function restartDaemonSet(name: string, ns: string = selectedNamespace.value) {
    isActionLoading.value = true
    try {
      await k8sApi.restartDaemonSet(ns, name)
      await fetchDaemonSets(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function fetchPods(ns: string = selectedNamespace.value) {
    try {
      pods.value = await k8sApi.getAllPods(ns)
    } catch (err) {
      logger.error('Failed to fetch pods', err)
    }
  }

  async function deletePod(name: string, ns: string = selectedNamespace.value) {
    isActionLoading.value = true
    try {
      await k8sApi.deletePod(ns, name)
      await fetchPods(ns)
    } finally {
      isActionLoading.value = false
    }
  }

  async function getResourceYAML(kind: string, name: string, ns: string = selectedNamespace.value): Promise<ResourceYAMLResponse> {
    return await k8sApi.getResourceYAML(kind, ns, name)
  }

  async function fetchPodMetrics(ns = selectedNamespace.value) {
    try {
      const list = (await k8sApi.getPodMetrics(ns)) || []
      const map: Record<string, PodMetrics> = {}
      if (Array.isArray(list)) {
        for (const m of list) {
          if (m && m.pod_name) {
            map[m.pod_name] = m
          }
        }
      }
      podMetrics.value = map
    } catch (err) {
      logger.error('Failed to fetch pod metrics', err)
      podMetrics.value = {}
    }
  }

  async function getServiceEndpoints(ns: string, name: string) {
    return await k8sApi.getServiceEndpoints(ns, name)
  }

  async function createNamespace(payload: CreateNamespacePayload) {
    isActionLoading.value = true
    try {
      const ns = await k8sApi.createNamespace(payload)
      await fetchNamespaces()
      return ns
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to create namespace')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  async function deleteNamespace(name: string) {
    isActionLoading.value = true
    try {
      await k8sApi.deleteNamespace(name)
      await fetchNamespaces()
      if (selectedNamespace.value === name) {
        setNamespace('default')
      }
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to delete namespace')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  async function fetchResourceQuotas(ns = selectedNamespace.value) {
    try {
      const list = await k8sApi.getResourceQuotas(ns)
      resourceQuotas.value = Array.isArray(list) ? list : []
    } catch (err) {
      logger.error('Failed to fetch resource quotas', err)
      resourceQuotas.value = []
    }
  }

  async function fetchEventsFeed(ns?: string, eventType?: string, kind?: string) {
    try {
      const list = await k8sApi.getClusterEventsFeed(ns, eventType, kind)
      eventsFeed.value = Array.isArray(list) ? list : []
    } catch (err) {
      logger.error('Failed to fetch events feed', err)
      eventsFeed.value = []
    }
  }

  return {
    clusterInfo,
    clusterOverview,
    nodes,
    namespaces,
    selectedNamespace,
    secrets,
    configmaps,
    deployments,
    statefulsets,
    daemonsets,
    pods,
    podMetrics,
    resourceQuotas,
    eventsFeed,
    services,
    ingresses,
    cronjobs,
    events,
    pvcs,
    pvs,
    isLoading,
    isActionLoading,
    error,
    setNamespace,
    fetchClusterInfo,
    fetchClusterOverview,
    fetchNodes,
    fetchNamespaces,
    fetchSecrets,
    fetchConfigMaps,
    fetchDeployments,
    fetchStatefulSets,
    scaleStatefulSet,
    restartStatefulSet,
    fetchDaemonSets,
    restartDaemonSet,
    fetchPods,
    deletePod,
    getResourceYAML,
    fetchServices,
    fetchIngresses,
    fetchCronJobs,
    fetchEvents,
    fetchPVCs,
    fetchPVs,
    fetchAllResources,
    getSecretDetail,
    saveSecret,
    deleteSecret,
    getConfigMapDetail,
    saveConfigMap,
    deleteConfigMap,
    restartDeployment,
    getDeploymentDetail,
    updateDeployment,
    scaleDeployment,
    getDeploymentPods,
    getPodLogs,
    getServiceDetail,
    getIngressDetail,
    getCronJobDetail,
    createCronJob,
    updateCronJob,
    toggleSuspendCronJob,
    triggerCronJobNow,
    getCronJobJobs,
    deleteCronJob,
    applyYAML,
    fetchPodMetrics,
    getServiceEndpoints,
    createNamespace,
    deleteNamespace,
    fetchResourceQuotas,
    fetchEventsFeed,
  }
})
