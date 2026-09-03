package k8smodule

import "time"

// ClusterInfoDTO holds cluster connection metadata and aggregate counts.
type ClusterInfoDTO struct {
	Connected       bool   `json:"connected"`
	Endpoint        string `json:"endpoint"`
	ServerVersion   string `json:"server_version"`
	CurrentContext  string `json:"current_context"`
	NamespaceCount  int    `json:"namespace_count"`
	SecretCount     int    `json:"secret_count"`
	ConfigMapCount  int    `json:"configmap_count"`
	DeploymentCount int    `json:"deployment_count"`
}

// NamespaceDTO represents a Kubernetes namespace.
type NamespaceDTO struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Age       string    `json:"age"`
}

// SecretItemDTO represents a summary item in the secrets list.
type SecretItemDTO struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	KeyCount  int       `json:"key_count"`
	Keys      []string  `json:"keys"`
	CreatedAt time.Time `json:"created_at"`
	Age       string    `json:"age"`
}

// SecretDetailDTO represents full secret details with decoded plaintext data.
type SecretDetailDTO struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Type            string            `json:"type"`
	Data            map[string]string `json:"data"`     // decoded plaintext values
	RawData         map[string]string `json:"raw_data"` // base64 values
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	ResourceVersion string            `json:"resource_version"`
	UID             string            `json:"uid"`
	CreatedAt       time.Time         `json:"created_at"`
}

// SaveSecretRequest payload for creating or updating a secret.
type SaveSecretRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=253"`
	Namespace   string            `json:"namespace" validate:"required"`
	Type        string            `json:"type"`
	Data        map[string]string `json:"data" validate:"required"` // plaintext key-value pairs
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ConfigMapItemDTO represents a summary item in the configmaps list.
type ConfigMapItemDTO struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	KeyCount  int       `json:"key_count"`
	Keys      []string  `json:"keys"`
	CreatedAt time.Time `json:"created_at"`
	Age       string    `json:"age"`
}

// ConfigMapDetailDTO represents full configmap details.
type ConfigMapDetailDTO struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Data            map[string]string `json:"data"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	ResourceVersion string            `json:"resource_version"`
	UID             string            `json:"uid"`
	CreatedAt       time.Time         `json:"created_at"`
}

// SaveConfigMapRequest payload for creating or updating a configmap.
type SaveConfigMapRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=253"`
	Namespace   string            `json:"namespace" validate:"required"`
	Data        map[string]string `json:"data" validate:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// DeploymentItemDTO represents a deployment workload.
type DeploymentItemDTO struct {
	Name          string    `json:"name"`
	Namespace     string    `json:"namespace"`
	Replicas      int32     `json:"replicas"`
	ReadyReplicas int32     `json:"ready_replicas"`
	Images        []string  `json:"images"`
	EnvSecrets    []string  `json:"env_secrets"`    // names of secrets used in env/envFrom
	EnvConfigMaps []string  `json:"env_configmaps"` // names of configmaps used in env/envFrom
	CreatedAt     time.Time `json:"created_at"`
	Age           string    `json:"age"`
}

// RolloutRestartResponse holds the response after restarting a deployment.
type RolloutRestartResponse struct {
	Message    string `json:"message"`
	Deployment string `json:"deployment"`
	Namespace  string `json:"namespace"`
	RestartAt  string `json:"restart_at"`
}

// ContainerEnvVarDTO represents an environment variable in a container.
type ContainerEnvVarDTO struct {
	Name      string `json:"name"`
	Value     string `json:"value,omitempty"`
	SecretRef string `json:"secret_ref,omitempty"`
	ConfigRef string `json:"config_ref,omitempty"`
}

// ContainerEnvFromDTO represents an envFrom source in a container.
type ContainerEnvFromDTO struct {
	Type   string `json:"type"` // "secret" or "configMap"
	Name   string `json:"name"`
	Prefix string `json:"prefix,omitempty"`
}

// ContainerDetailDTO represents a container inside a pod or deployment.
type ContainerDetailDTO struct {
	Name    string                `json:"name"`
	Image   string                `json:"image"`
	Env     []ContainerEnvVarDTO  `json:"env"`
	EnvFrom []ContainerEnvFromDTO `json:"env_from"`
}

// DeploymentDetailDTO holds full deployment configuration.
type DeploymentDetailDTO struct {
	Name          string               `json:"name"`
	Namespace     string               `json:"namespace"`
	Replicas      int32                `json:"replicas"`
	ReadyReplicas int32                `json:"ready_replicas"`
	Labels        map[string]string    `json:"labels"`
	Annotations   map[string]string    `json:"annotations"`
	Containers    []ContainerDetailDTO `json:"containers"`
	CreatedAt     time.Time            `json:"created_at"`
	Age           string               `json:"age"`
}

// UpdateDeploymentRequest holds update payload for a deployment.
type UpdateDeploymentRequest struct {
	Replicas   *int32               `json:"replicas"`
	Containers []ContainerDetailDTO `json:"containers"`
}

// PodItemDTO represents a pod belonging to a deployment or namespace.
type PodItemDTO struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Phase        string    `json:"phase"` // Running, Pending, Succeeded, Failed, CrashLoopBackOff
	StatusReason string    `json:"status_reason,omitempty"`
	Ready        string    `json:"ready"` // e.g. "1/1"
	Restarts     int32     `json:"restarts"`
	Node         string    `json:"node"`
	IP           string    `json:"ip"`
	Containers   []string  `json:"containers"`
	CreatedAt    time.Time `json:"created_at"`
	Age          string    `json:"age"`
}

// PodLogsResponseDTO holds container logs.
type PodLogsResponseDTO struct {
	Pod       string `json:"pod"`
	Container string `json:"container"`
	Namespace string `json:"namespace"`
	Logs      string `json:"logs"`
	LineCount int    `json:"line_count"`
}

// ServicePortDTO represents a single port mapping in a Kubernetes Service.
type ServicePortDTO struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
	NodePort   int32  `json:"node_port,omitempty"`
}

// ServiceItemDTO represents a Kubernetes Service in a list.
type ServiceItemDTO struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"` // ClusterIP, NodePort, LoadBalancer, ExternalName
	ClusterIP  string            `json:"cluster_ip"`
	ExternalIP string            `json:"external_ip,omitempty"`
	Ports      []ServicePortDTO  `json:"ports"`
	Selector   map[string]string `json:"selector"`
	CreatedAt  time.Time         `json:"created_at"`
	Age        string            `json:"age"`
}

// ServiceDetailDTO represents detailed information of a Kubernetes Service.
type ServiceDetailDTO struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"cluster_ip"`
	ExternalIP  string            `json:"external_ip,omitempty"`
	Ports       []ServicePortDTO  `json:"ports"`
	Selector    map[string]string `json:"selector"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	Age         string            `json:"age"`
}

// IngressRuleDTO represents a path route rule in an Ingress.
type IngressRuleDTO struct {
	Host        string `json:"host"`
	Path        string `json:"path"`
	PathType    string `json:"path_type"`
	ServiceName string `json:"service_name"`
	ServicePort int32  `json:"service_port"`
}

// IngressItemDTO represents a Kubernetes Ingress resource.
type IngressItemDTO struct {
	Name        string           `json:"name"`
	Namespace   string           `json:"namespace"`
	ClassName   string           `json:"class_name"`
	Hosts       []string         `json:"hosts"`
	Address     string           `json:"address"`
	Ports       []string         `json:"ports"`
	TLS         []string         `json:"tls"`
	Rules       []IngressRuleDTO `json:"rules"`
	CreatedAt   time.Time        `json:"created_at"`
	Age         string           `json:"age"`
}

// CronJobItemDTO represents a Kubernetes CronJob in a list.
type CronJobItemDTO struct {
	Name             string     `json:"name"`
	Namespace        string     `json:"namespace"`
	Schedule         string     `json:"schedule"`
	Suspend          bool       `json:"suspend"`
	ActiveJobs       int        `json:"active_jobs"`
	LastScheduleTime *time.Time `json:"last_schedule_time,omitempty"`
	Image            string     `json:"image"`
	CreatedAt        time.Time  `json:"created_at"`
	Age              string     `json:"age"`
}

// CronJobDetailDTO represents detailed information of a Kubernetes CronJob.
type CronJobDetailDTO struct {
	Name                       string               `json:"name"`
	Namespace                  string               `json:"namespace"`
	Schedule                   string               `json:"schedule"`
	Suspend                    bool                 `json:"suspend"`
	ConcurrencyPolicy          string               `json:"concurrency_policy"`
	SuccessfulJobsHistoryLimit *int32               `json:"successful_jobs_history_limit,omitempty"`
	FailedJobsHistoryLimit     *int32               `json:"failed_jobs_history_limit,omitempty"`
	Containers                 []ContainerDetailDTO `json:"containers"`
	Labels                     map[string]string    `json:"labels"`
	Annotations                map[string]string    `json:"annotations"`
	LastScheduleTime           *time.Time           `json:"last_schedule_time,omitempty"`
	CreatedAt                  time.Time            `json:"created_at"`
	Age                        string               `json:"age"`
}

// JobItemDTO represents a Kubernetes Job execution instance.
type JobItemDTO struct {
	Name           string     `json:"name"`
	Namespace      string     `json:"namespace"`
	CronJobName    string     `json:"cron_job_name,omitempty"`
	Status         string     `json:"status"` // Complete, Running, Failed
	Succeeded      int32      `json:"succeeded"`
	Failed         int32      `json:"failed"`
	Active         int32      `json:"active"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	CompletionTime *time.Time `json:"completion_time,omitempty"`
	Duration       string     `json:"duration"`
	CreatedAt      time.Time  `json:"created_at"`
	Age            string     `json:"age"`
}

// CreateCronJobRequest represents parameters to create a new CronJob.
type CreateCronJobRequest struct {
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Schedule   string                 `json:"schedule"`
	Suspend    bool                   `json:"suspend"`
	Containers []ContainerDetailDTO   `json:"containers"`
	Labels     map[string]string      `json:"labels,omitempty"`
}

// UpdateCronJobRequest represents parameters to update an existing CronJob.
type UpdateCronJobRequest struct {
	Schedule   string               `json:"schedule"`
	Suspend    *bool                `json:"suspend,omitempty"`
	Containers []ContainerDetailDTO `json:"containers"`
}

// ScaleDeploymentRequest represents the request body to scale replicas.
type ScaleDeploymentRequest struct {
	Replicas int32 `json:"replicas"`
}

// EventItemDTO represents a Kubernetes Event.
type EventItemDTO struct {
	Type           string    `json:"type"` // Normal, Warning
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	InvolvedObject string    `json:"involved_object"` // Kind/Name
	Count          int32     `json:"count"`
	FirstTime      time.Time `json:"first_time"`
	LastTime       time.Time `json:"last_time"`
	Age            string    `json:"age"`
}

// PVCItemDTO represents a PersistentVolumeClaim.
type PVCItemDTO struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Status       string    `json:"status"` // Bound, Pending, Lost
	Volume       string    `json:"volume"`
	Capacity     string    `json:"capacity"`
	StorageClass string    `json:"storage_class"`
	AccessModes  []string  `json:"access_modes"`
	Age          string    `json:"age"`
	CreatedAt    time.Time `json:"created_at"`
}

// PVItemDTO represents a PersistentVolume.
type PVItemDTO struct {
	Name          string    `json:"name"`
	Capacity      string    `json:"capacity"`
	AccessModes   []string  `json:"access_modes"`
	ReclaimPolicy string    `json:"reclaim_policy"`
	Status        string    `json:"status"`
	Claim         string    `json:"claim"`
	StorageClass  string    `json:"storage_class"`
	Age           string    `json:"age"`
	CreatedAt     time.Time `json:"created_at"`
}

// NodeConditionDTO represents a single condition on a Kubernetes Node.
type NodeConditionDTO struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// NodeDTO represents a node in the Kubernetes cluster.
type NodeDTO struct {
	Name              string             `json:"name"`
	Status            string             `json:"status"` // Ready, NotReady
	Roles             []string           `json:"roles"`
	Version           string             `json:"version"`
	OSImage           string             `json:"os_image"`
	KernelVersion     string             `json:"kernel_version"`
	ContainerRuntime  string             `json:"container_runtime"`
	InternalIP        string             `json:"internal_ip"`
	ExternalIP        string             `json:"external_ip"`
	CPUCapacity       string             `json:"cpu_capacity"`
	CPUAllocatable    string             `json:"cpu_allocatable"`
	MemoryCapacity    string             `json:"memory_capacity"`
	MemoryAllocatable string             `json:"memory_allocatable"`
	PodsCapacity      int64              `json:"pods_capacity"`
	PodsAllocatable   int64              `json:"pods_allocatable"`
	Conditions        []NodeConditionDTO `json:"conditions"`
	CreatedAt         time.Time          `json:"created_at"`
	Age               string             `json:"age"`
}

// ClusterOverviewDTO provides aggregated telemetry and inventory counts for the cluster.
type ClusterOverviewDTO struct {
	NodesReady          int            `json:"nodes_ready"`
	NodesTotal          int            `json:"nodes_total"`
	TotalCPUCores       float64        `json:"total_cpu_cores"`
	AllocatableCPUCores float64        `json:"allocatable_cpu_cores"`
	TotalMemoryGiB      float64        `json:"total_memory_gib"`
	AllocatableMemoryGiB float64       `json:"allocatable_memory_gib"`
	TotalPodsCapacity   int64          `json:"total_pods_capacity"`
	ActivePodsCount     int            `json:"active_pods_count"`
	DeploymentsCount    int            `json:"deployments_count"`
	StatefulSetsCount   int            `json:"statefulsets_count"`
	DaemonSetsCount     int            `json:"daemonsets_count"`
	ServicesCount       int            `json:"services_count"`
	IngressesCount      int            `json:"ingresses_count"`
	PVCsCount           int            `json:"pvcs_count"`
	PVsCount            int            `json:"pvs_count"`
	NamespacesCount     int            `json:"namespaces_count"`
	CronJobsCount       int            `json:"cronjobs_count"`
	WarningEvents       []EventItemDTO `json:"warning_events"`
	Nodes               []NodeDTO      `json:"nodes"`
}

// StatefulSetItemDTO represents a Kubernetes StatefulSet.
type StatefulSetItemDTO struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Replicas        int32             `json:"replicas"`
	ReadyReplicas   int32             `json:"ready_replicas"`
	CurrentReplicas int32             `json:"current_replicas"`
	Images          []string          `json:"images"`
	Labels          map[string]string `json:"labels"`
	CreatedAt       time.Time         `json:"created_at"`
	Age             string            `json:"age"`
}

// ScaleStatefulSetRequest holds payload for scaling a StatefulSet.
type ScaleStatefulSetRequest struct {
	Replicas int32 `json:"replicas"`
}

// DaemonSetItemDTO represents a Kubernetes DaemonSet.
type DaemonSetItemDTO struct {
	Name                   string            `json:"name"`
	Namespace              string            `json:"namespace"`
	DesiredNumberScheduled int32             `json:"desired_number_scheduled"`
	CurrentNumberScheduled int32             `json:"current_number_scheduled"`
	NumberReady            int32             `json:"number_ready"`
	NumberAvailable        int32             `json:"number_available"`
	Images                 []string          `json:"images"`
	Labels                 map[string]string `json:"labels"`
	CreatedAt              time.Time         `json:"created_at"`
	Age                    string            `json:"age"`
}

// ResourceYAMLResponseDTO holds live YAML serialization of any resource.
type ResourceYAMLResponseDTO struct {
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	APIVersion string `json:"api_version"`
	YAML       string `json:"yaml"`
}

// EndpointTargetDTO represents a specific backend Pod address behind a Service.
type EndpointTargetDTO struct {
	IP       string `json:"ip"`
	PodName  string `json:"pod_name"`
	NodeName string `json:"node_name"`
	Ready    bool   `json:"ready"`
}

// ServiceEndpointPortDTO represents a port exposed by an endpoint.
type ServiceEndpointPortDTO struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// ServiceEndpointsDTO represents all resolved network endpoints behind a Kubernetes Service.
type ServiceEndpointsDTO struct {
	ServiceName string                   `json:"service_name"`
	Namespace   string                   `json:"namespace"`
	Ports       []ServiceEndpointPortDTO `json:"ports"`
	Targets     []EndpointTargetDTO      `json:"targets"`
}

// PodMetricsDTO holds actual/live CPU and Memory resource consumption for a Pod.
type PodMetricsDTO struct {
	PodName       string  `json:"pod_name"`
	Namespace     string  `json:"namespace"`
	CPUUsage      string  `json:"cpu_usage"`      // e.g. "15m"
	MemoryUsage   string  `json:"memory_usage"`   // e.g. "64Mi"
	CPUPercent    float64 `json:"cpu_percent"`    // e.g. 12.5%
	MemoryPercent float64 `json:"memory_percent"` // e.g. 24.8%
}

// CreateNamespaceRequest holds payload to create a new Kubernetes Namespace.
type CreateNamespaceRequest struct {
	Name   string            `json:"name" validate:"required"`
	Labels map[string]string `json:"labels,omitempty"`
}

// ResourceQuotaItemDTO represents a namespace resource quota.
type ResourceQuotaItemDTO struct {
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	CPULimit    string    `json:"cpu_limit"`
	CPUUsed     string    `json:"cpu_used"`
	MemoryLimit string    `json:"memory_limit"`
	MemoryUsed  string    `json:"memory_used"`
	PodsLimit   string    `json:"pods_limit"`
	PodsUsed    string    `json:"pods_used"`
	Age         string    `json:"age"`
	CreatedAt   time.Time `json:"created_at"`
}


