export interface ClusterInfo {
  connected: boolean
  endpoint: string
  server_version: string
  current_context: string
  namespace_count: number
  secret_count: number
  configmap_count: number
  deployment_count: number
}

export interface NamespaceItem {
  name: string
  status: string
  created_at: string
  age: string
}

export interface SecretItem {
  name: string
  namespace: string
  type: string
  key_count: number
  keys: string[]
  created_at: string
  age: string
}

export interface SecretDetail {
  name: string
  namespace: string
  type: string
  data: Record<string, string> // decoded plaintext values
  raw_data?: Record<string, string> // base64 values
  labels?: Record<string, string>
  annotations?: Record<string, string>
  resource_version?: string
  uid?: string
  created_at?: string
}

export interface SaveSecretPayload {
  name: string
  namespace: string
  type?: string
  data: Record<string, string>
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

export interface ConfigMapItem {
  name: string
  namespace: string
  key_count: number
  keys: string[]
  created_at: string
  age: string
}

export interface ConfigMapDetail {
  name: string
  namespace: string
  type?: string
  data: Record<string, string>
  labels?: Record<string, string>
  annotations?: Record<string, string>
  resource_version?: string
  uid?: string
  created_at?: string
}

export interface SaveConfigMapPayload {
  name: string
  namespace: string
  data: Record<string, string>
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

export interface DeploymentItem {
  name: string
  namespace: string
  replicas: number
  ready_replicas: number
  images: string[]
  env_secrets: string[]
  env_configmaps: string[]
  created_at: string
  age: string
}

export interface RolloutRestartResponse {
  message: string
  deployment: string
  namespace: string
  restart_at: string
}

export interface ContainerEnvVar {
  name: string
  value?: string
  secret_ref?: string
  config_ref?: string
}

export interface ContainerEnvFrom {
  type: string
  name: string
  prefix?: string
}

export interface ContainerDetail {
  name: string
  image: string
  env: ContainerEnvVar[]
  env_from: ContainerEnvFrom[]
}

export interface DeploymentDetail {
  name: string
  namespace: string
  replicas: number
  ready_replicas: number
  labels?: Record<string, string>
  annotations?: Record<string, string>
  containers: ContainerDetail[]
  created_at: string
  age: string
}

export interface UpdateDeploymentPayload {
  replicas?: number
  containers?: ContainerDetail[]
}

export interface PodItem {
  name: string
  namespace: string
  phase: string
  status_reason?: string
  ready: string
  restarts: number
  node: string
  ip: string
  containers: string[]
  created_at: string
  age: string
}

export interface PodLogsResponse {
  pod: string
  container: string
  namespace: string
  logs: string
  line_count: number
}

export interface ServicePort {
  name: string
  port: number
  target_port: string
  protocol: string
  node_port?: number
}

export interface ServiceItem {
  name: string
  namespace: string
  type: string
  cluster_ip: string
  external_ip?: string
  ports: ServicePort[]
  selector?: Record<string, string>
  created_at: string
  age: string
}

export interface ServiceDetail extends ServiceItem {
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

export interface IngressRule {
  host: string
  path: string
  path_type: string
  service_name: string
  service_port: number
}

export interface IngressItem {
  name: string
  namespace: string
  class_name: string
  hosts: string[]
  address: string
  ports: string[]
  tls: string[]
  rules: IngressRule[]
  created_at: string
  age: string
}

export interface CronJobItem {
  name: string
  namespace: string
  schedule: string
  suspend: boolean
  active_jobs: number
  last_schedule_time?: string
  image: string
  created_at: string
  age: string
}

export interface CronJobDetail {
  name: string
  namespace: string
  schedule: string
  suspend: boolean
  concurrency_policy: string
  successful_jobs_history_limit?: number
  failed_jobs_history_limit?: number
  containers: ContainerDetail[]
  labels?: Record<string, string>
  annotations?: Record<string, string>
  last_schedule_time?: string
  created_at: string
  age: string
}

export interface JobItem {
  name: string
  namespace: string
  cron_job_name?: string
  status: 'Complete' | 'Running' | 'Failed'
  succeeded: number
  failed: number
  active: number
  start_time?: string
  completion_time?: string
  duration: string
  created_at: string
  age: string
}

export interface CreateCronJobPayload {
  name: string
  namespace: string
  schedule: string
  suspend: boolean
  containers: ContainerDetail[]
  labels?: Record<string, string>
}

export interface UpdateCronJobPayload {
  schedule: string
  suspend?: boolean
  containers: ContainerDetail[]
}

export interface EventItem {
  type: string
  reason: string
  message: string
  involved_object: string
  count: number
  first_time: string
  last_time: string
  age: string
  namespace?: string
}

export interface PVCItem {
  name: string
  namespace: string
  status: string
  volume: string
  capacity: string
  storage_class: string
  access_modes: string[]
  age: string
  created_at: string
}

export interface PVItem {
  name: string
  capacity: string
  access_modes: string[]
  reclaim_policy: string
  status: string
  claim: string
  storage_class: string
  age: string
  created_at: string
}

export interface NodeCondition {
  type: string
  status: string
  reason?: string
  message?: string
}

export interface NodeItem {
  name: string
  status: string
  roles: string[]
  version: string
  os_image: string
  kernel_version: string
  container_runtime: string
  internal_ip: string
  external_ip: string
  cpu_capacity: string
  cpu_allocatable: string
  memory_capacity: string
  memory_allocatable: string
  pods_capacity: number
  pods_allocatable: number
  conditions: NodeCondition[]
  created_at: string
  age: string
}

export interface ClusterOverview {
  nodes_ready: number
  nodes_total: number
  total_cpu_cores: number
  allocatable_cpu_cores: number
  total_memory_gib: number
  allocatable_memory_gib: number
  total_pods_capacity: number
  active_pods_count: number
  deployments_count: number
  statefulsets_count: number
  daemonsets_count: number
  services_count: number
  ingresses_count: number
  pvcs_count: number
  pvs_count: number
  namespaces_count: number
  cronjobs_count: number
  warning_events: EventItem[]
  nodes: NodeItem[]
}

export interface StatefulSetItem {
  name: string
  namespace: string
  replicas: number
  ready_replicas: number
  current_replicas: number
  images: string[]
  labels?: Record<string, string>
  created_at: string
  age: string
}

export interface DaemonSetItem {
  name: string
  namespace: string
  desired_number_scheduled: number
  current_number_scheduled: number
  number_ready: number
  number_available: number
  images: string[]
  labels?: Record<string, string>
  created_at: string
  age: string
}

export interface ResourceYAMLResponse {
  kind: string
  namespace: string
  name: string
  api_version: string
  yaml: string
}

export interface EndpointTarget {
  ip: string
  pod_name: string
  node_name: string
  ready: boolean
}

export interface ServiceEndpointPort {
  name: string
  port: number
  protocol: string
}

export interface ServiceEndpoints {
  service_name: string
  namespace: string
  ports: ServiceEndpointPort[]
  targets: EndpointTarget[]
}

export interface PodMetrics {
  pod_name: string
  namespace: string
  cpu_usage: string
  memory_usage: string
  cpu_percent: number
  memory_percent: number
}

export interface CreateNamespacePayload {
  name: string
  labels?: Record<string, string>
}

export interface ResourceQuotaItem {
  name: string
  namespace: string
  cpu_limit: string
  cpu_used: string
  memory_limit: string
  memory_used: string
  pods_limit: string
  pods_used: string
  age: string
  created_at: string
}



