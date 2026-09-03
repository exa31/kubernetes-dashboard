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
  ServiceEndpoints,
  ServiceItem,
  StatefulSetItem,
  UpdateCronJobPayload,
  UpdateDeploymentPayload
} from '@/types'

import { apiClient } from './client'

export const k8sApi = {
  getClusterInfo: async (): Promise<ClusterInfo> => {
    const res = await apiClient.get<{ data: ClusterInfo }>('/k8s/cluster-info')
    return res.data.data
  },

  getNamespaces: async (): Promise<NamespaceItem[]> => {
    const res = await apiClient.get<{ data: NamespaceItem[] }>('/k8s/namespaces')
    return res.data.data
  },

  getSecrets: async (namespace: string): Promise<SecretItem[]> => {
    const res = await apiClient.get<{ data: SecretItem[] }>(
      `/k8s/secrets?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getSecret: async (namespace: string, name: string): Promise<SecretDetail> => {
    const res = await apiClient.get<{ data: SecretDetail }>(
      `/k8s/secrets/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  saveSecret: async (payload: SaveSecretPayload): Promise<SecretDetail> => {
    const res = await apiClient.post<{ data: SecretDetail }>('/k8s/secrets', payload)
    return res.data.data
  },

  deleteSecret: async (namespace: string, name: string): Promise<{ deleted: boolean }> => {
    const res = await apiClient.delete<{ data: { deleted: boolean } }>(
      `/k8s/secrets/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  getConfigMaps: async (namespace: string): Promise<ConfigMapItem[]> => {
    const res = await apiClient.get<{ data: ConfigMapItem[] }>(
      `/k8s/configmaps?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getConfigMap: async (namespace: string, name: string): Promise<ConfigMapDetail> => {
    const res = await apiClient.get<{ data: ConfigMapDetail }>(
      `/k8s/configmaps/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  saveConfigMap: async (payload: SaveConfigMapPayload): Promise<ConfigMapDetail> => {
    const res = await apiClient.post<{ data: ConfigMapDetail }>('/k8s/configmaps', payload)
    return res.data.data
  },

  deleteConfigMap: async (namespace: string, name: string): Promise<{ deleted: boolean }> => {
    const res = await apiClient.delete<{ data: { deleted: boolean } }>(
      `/k8s/configmaps/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  getDeployments: async (namespace: string): Promise<DeploymentItem[]> => {
    const res = await apiClient.get<{ data: DeploymentItem[] }>(
      `/k8s/deployments?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  restartDeployment: async (namespace: string, name: string): Promise<RolloutRestartResponse> => {
    const res = await apiClient.post<{ data: RolloutRestartResponse }>(
      `/k8s/deployments/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/restart`
    )
    return res.data.data
  },

  getDeploymentDetail: async (namespace: string, name: string): Promise<DeploymentDetail> => {
    const res = await apiClient.get<{ data: DeploymentDetail }>(
      `/k8s/deployments/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  updateDeployment: async (
    namespace: string,
    name: string,
    payload: UpdateDeploymentPayload
  ): Promise<DeploymentDetail> => {
    const res = await apiClient.put<{ data: DeploymentDetail }>(
      `/k8s/deployments/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`,
      payload
    )
    return res.data.data
  },

  getDeploymentPods: async (namespace: string, name: string): Promise<PodItem[]> => {
    const res = await apiClient.get<{ data: PodItem[] }>(
      `/k8s/deployments/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/pods`
    )
    return res.data.data
  },

  getPodLogs: async (
    namespace: string,
    podName: string,
    params?: { container?: string; tail_lines?: number; timestamps?: boolean }
  ): Promise<PodLogsResponse> => {
    const query = new URLSearchParams()
    if (params?.container) query.set('container', params.container)
    if (params?.tail_lines) query.set('tail_lines', String(params.tail_lines))
    if (params?.timestamps !== undefined) query.set('timestamps', String(params.timestamps))

    const qStr = query.toString()
    const url = `/k8s/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(podName)}/logs${qStr ? `?${qStr}` : ''}`
    const res = await apiClient.get<{ data: PodLogsResponse }>(url)
    return res.data.data
  },

  getServices: async (namespace: string): Promise<ServiceItem[]> => {
    const res = await apiClient.get<{ data: ServiceItem[] }>(
      `/k8s/services?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getService: async (namespace: string, name: string): Promise<ServiceDetail> => {
    const res = await apiClient.get<{ data: ServiceDetail }>(
      `/k8s/services/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  getIngresses: async (namespace: string): Promise<IngressItem[]> => {
    const res = await apiClient.get<{ data: IngressItem[] }>(
      `/k8s/ingresses?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getIngress: async (namespace: string, name: string): Promise<IngressItem> => {
    const res = await apiClient.get<{ data: IngressItem }>(
      `/k8s/ingresses/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  getCronJobs: async (namespace: string): Promise<CronJobItem[]> => {
    const res = await apiClient.get<{ data: CronJobItem[] }>(
      `/k8s/cronjobs?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getCronJob: async (namespace: string, name: string): Promise<CronJobDetail> => {
    const res = await apiClient.get<{ data: CronJobDetail }>(
      `/k8s/cronjobs/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  createCronJob: async (payload: CreateCronJobPayload): Promise<CronJobDetail> => {
    const res = await apiClient.post<{ data: CronJobDetail }>('/k8s/cronjobs', payload)
    return res.data.data
  },

  updateCronJob: async (
    namespace: string,
    name: string,
    payload: UpdateCronJobPayload
  ): Promise<CronJobDetail> => {
    const res = await apiClient.put<{ data: CronJobDetail }>(
      `/k8s/cronjobs/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`,
      payload
    )
    return res.data.data
  },

  toggleSuspendCronJob: async (
    namespace: string,
    name: string
  ): Promise<{ suspended: boolean }> => {
    const res = await apiClient.post<{ data: { suspended: boolean } }>(
      `/k8s/cronjobs/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/toggle-suspend`
    )
    return res.data.data
  },

  triggerCronJobNow: async (namespace: string, name: string): Promise<JobItem> => {
    const res = await apiClient.post<{ data: JobItem }>(
      `/k8s/cronjobs/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/run`
    )
    return res.data.data
  },

  getCronJobJobs: async (namespace: string, name: string): Promise<JobItem[]> => {
    const res = await apiClient.get<{ data: JobItem[] }>(
      `/k8s/cronjobs/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/jobs`
    )
    return res.data.data
  },

  deleteCronJob: async (namespace: string, name: string): Promise<void> => {
    await apiClient.delete(
      `/k8s/cronjobs/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`
    )
  },

  scaleDeployment: async (
    namespace: string,
    name: string,
    replicas: number
  ): Promise<DeploymentDetail> => {
    const res = await apiClient.put<{ data: DeploymentDetail }>(
      `/k8s/deployments/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/scale`,
      { replicas }
    )
    return res.data.data
  },

  getEvents: async (namespace: string): Promise<EventItem[]> => {
    const res = await apiClient.get<{ data: EventItem[] }>(
      `/k8s/events?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getPVCs: async (namespace: string): Promise<PVCItem[]> => {
    const res = await apiClient.get<{ data: PVCItem[] }>(
      `/k8s/pvcs?namespace=${encodeURIComponent(namespace)}`
    )
    return res.data.data
  },

  getPVs: async (): Promise<PVItem[]> => {
    const res = await apiClient.get<{ data: PVItem[] }>('/k8s/pvs')
    return res.data.data
  },

  applyYAML: async (yaml: string, namespace?: string, dryRun = false): Promise<ApplyYAMLResult> => {
    const res = await apiClient.post<{ data: ApplyYAMLResult }>('/k8s/apply-yaml', {
      yaml,
      namespace: namespace || undefined,
      dry_run: dryRun
    })
    return res.data.data
  },

  getClusterOverview: async (): Promise<ClusterOverview> => {
    const res = await apiClient.get<{ data: ClusterOverview }>('/k8s/cluster-overview')
    return res.data.data
  },

  getNodes: async (): Promise<NodeItem[]> => {
    const res = await apiClient.get<{ data: NodeItem[] }>('/k8s/nodes')
    return res.data.data ?? []
  },

  getStatefulSets: async (namespace?: string): Promise<StatefulSetItem[]> => {
    const url = namespace
      ? `/k8s/statefulsets?namespace=${encodeURIComponent(namespace)}`
      : '/k8s/statefulsets'
    const res = await apiClient.get<{ data: StatefulSetItem[] }>(url)
    return res.data.data ?? []
  },

  scaleStatefulSet: async (namespace: string, name: string, replicas: number): Promise<void> => {
    await apiClient.put(
      `/k8s/statefulsets/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/scale`,
      {
        replicas
      }
    )
  },

  restartStatefulSet: async (namespace: string, name: string): Promise<void> => {
    await apiClient.post(
      `/k8s/statefulsets/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/restart`
    )
  },

  getDaemonSets: async (namespace?: string): Promise<DaemonSetItem[]> => {
    const url = namespace
      ? `/k8s/daemonsets?namespace=${encodeURIComponent(namespace)}`
      : '/k8s/daemonsets'
    const res = await apiClient.get<{ data: DaemonSetItem[] }>(url)
    return res.data.data ?? []
  },

  restartDaemonSet: async (namespace: string, name: string): Promise<void> => {
    await apiClient.post(
      `/k8s/daemonsets/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/restart`
    )
  },

  getAllPods: async (namespace?: string): Promise<PodItem[]> => {
    const url = namespace ? `/k8s/pods?namespace=${encodeURIComponent(namespace)}` : '/k8s/pods'
    const res = await apiClient.get<{ data: PodItem[] }>(url)
    return res.data.data ?? []
  },

  deletePod: async (namespace: string, name: string): Promise<void> => {
    await apiClient.delete(`/k8s/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`)
  },

  getResourceYAML: async (
    kind: string,
    namespace: string,
    name: string
  ): Promise<ResourceYAMLResponse> => {
    const res = await apiClient.get<{ data: ResourceYAMLResponse }>(
      `/k8s/resource-yaml?kind=${encodeURIComponent(kind)}&namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`
    )
    return res.data.data
  },

  getServiceEndpoints: async (namespace: string, name: string): Promise<ServiceEndpoints> => {
    const res = await apiClient.get<{ data: ServiceEndpoints }>(
      `/k8s/services/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/endpoints`
    )
    return res.data.data
  },

  getPodMetrics: async (namespace?: string): Promise<PodMetrics[]> => {
    const url = namespace
      ? `/k8s/metrics/pods?namespace=${encodeURIComponent(namespace)}`
      : '/k8s/metrics/pods'
    const res = await apiClient.get<{ data: PodMetrics[] }>(url)
    return res.data.data ?? []
  },

  createNamespace: async (payload: CreateNamespacePayload): Promise<NamespaceItem> => {
    const res = await apiClient.post<{ data: NamespaceItem }>('/k8s/namespaces', payload)
    return res.data.data
  },

  deleteNamespace: async (name: string): Promise<void> => {
    await apiClient.delete(`/k8s/namespaces/${encodeURIComponent(name)}`)
  },

  getResourceQuotas: async (namespace?: string): Promise<ResourceQuotaItem[]> => {
    const url = namespace
      ? `/k8s/resource-quotas?namespace=${encodeURIComponent(namespace)}`
      : '/k8s/resource-quotas'
    const res = await apiClient.get<{ data: ResourceQuotaItem[] }>(url)
    return res.data.data ?? []
  },

  getClusterEventsFeed: async (
    namespace?: string,
    eventType?: string,
    kind?: string
  ): Promise<EventItem[]> => {
    const params = new URLSearchParams()
    if (namespace) params.append('namespace', namespace)
    if (eventType) params.append('type', eventType)
    if (kind) params.append('kind', kind)
    const qs = params.toString() ? `?${params.toString()}` : ''
    const res = await apiClient.get<{ data: EventItem[] }>(`/k8s/events/feed${qs}`)
    return res.data.data ?? []
  }
}

export interface AppliedResourceResult {
  api_version: string
  kind: string
  namespace: string
  name: string
  action: string
  status: 'success' | 'error'
  message: string
}

export interface ApplyYAMLResult {
  total: number
  success_count: number
  error_count: number
  dry_run: boolean
  results: AppliedResourceResult[]
}
