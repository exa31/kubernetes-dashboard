package k8smodule

import (
	"context"
	"strconv"

	"golang/pkg/errors"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// K8sHandler exposes Kubernetes endpoints via Fiber.
type K8sHandler struct {
	service *K8sService
}

// NewK8sHandler creates a new handler.
func NewK8sHandler(service *K8sService) *K8sHandler {
	return &K8sHandler{service: service}
}

// GetClusterInfo handles GET /api/v1/k8s/cluster-info.
func (h *K8sHandler) GetClusterInfo() fiber.Handler {
	return func(c *fiber.Ctx) error {
		info, err := h.service.GetClusterInfo(c.Context())
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, info, "Cluster info retrieved successfully")
	}
}

// ListNamespaces handles GET /api/v1/k8s/namespaces.
func (h *K8sHandler) ListNamespaces() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespaces, err := h.service.ListNamespaces(c.Context())
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, namespaces, "Namespaces retrieved successfully")
	}
}

// ListSecrets handles GET /api/v1/k8s/secrets.
func (h *K8sHandler) ListSecrets() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "default")
		secrets, err := h.service.ListSecrets(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, secrets, "Secrets retrieved successfully")
	}
}

// GetSecret handles GET /api/v1/k8s/secrets/:namespace/:name.
func (h *K8sHandler) GetSecret() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		secret, err := h.service.GetSecret(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, secret, "Secret retrieved successfully")
	}
}

// SaveSecret handles POST /api/v1/k8s/secrets.
func (h *K8sHandler) SaveSecret() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SaveSecretRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		secret, err := h.service.SaveSecret(c.Context(), &req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, secret, "Secret saved successfully")
	}
}

// DeleteSecret handles DELETE /api/v1/k8s/secrets/:namespace/:name.
func (h *K8sHandler) DeleteSecret() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		if err := h.service.DeleteSecret(c.Context(), namespace, name); err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"deleted": true, "name": name, "namespace": namespace}, "Secret deleted successfully")
	}
}

// ListConfigMaps handles GET /api/v1/k8s/configmaps.
func (h *K8sHandler) ListConfigMaps() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "default")
		cms, err := h.service.ListConfigMaps(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, cms, "ConfigMaps retrieved successfully")
	}
}

// GetConfigMap handles GET /api/v1/k8s/configmaps/:namespace/:name.
func (h *K8sHandler) GetConfigMap() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		cm, err := h.service.GetConfigMap(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, cm, "ConfigMap retrieved successfully")
	}
}

// SaveConfigMap handles POST /api/v1/k8s/configmaps.
func (h *K8sHandler) SaveConfigMap() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SaveConfigMapRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		cm, err := h.service.SaveConfigMap(c.Context(), &req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, cm, "ConfigMap saved successfully")
	}
}

// DeleteConfigMap handles DELETE /api/v1/k8s/configmaps/:namespace/:name.
func (h *K8sHandler) DeleteConfigMap() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		if err := h.service.DeleteConfigMap(c.Context(), namespace, name); err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"deleted": true, "name": name, "namespace": namespace}, "ConfigMap deleted successfully")
	}
}

// ListDeployments handles GET /api/v1/k8s/deployments.
func (h *K8sHandler) ListDeployments() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "default")
		deployments, err := h.service.ListDeployments(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, deployments, "Deployments retrieved successfully")
	}
}

// RolloutRestartDeployment handles POST /api/v1/k8s/deployments/:namespace/:name/restart.
func (h *K8sHandler) RolloutRestartDeployment() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		res, err := h.service.RolloutRestartDeployment(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, res, res.Message)
	}
}

// GetDeployment handles GET /api/v1/k8s/deployments/:namespace/:name.
func (h *K8sHandler) GetDeployment() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		dep, err := h.service.GetDeployment(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, dep, "Deployment retrieved successfully")
	}
}

// UpdateDeployment handles PUT /api/v1/k8s/deployments/:namespace/:name.
func (h *K8sHandler) UpdateDeployment() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")

		var req UpdateDeploymentRequest
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		updated, err := h.service.UpdateDeployment(c.Context(), namespace, name, &req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, updated, "Deployment updated successfully")
	}
}

// GetDeploymentPods handles GET /api/v1/k8s/deployments/:namespace/:name/pods.
func (h *K8sHandler) GetDeploymentPods() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		pods, err := h.service.GetDeploymentPods(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, pods, "Deployment pods retrieved successfully")
	}
}

// GetPodLogs handles GET /api/v1/k8s/pods/:namespace/:name/logs.
func (h *K8sHandler) GetPodLogs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		podName := c.Params("name")
		container := c.Query("container", "")
		tailLinesStr := c.Query("tail_lines", "250")
		timestampsStr := c.Query("timestamps", "false")

		tailLines, _ := strconv.ParseInt(tailLinesStr, 10, 64)
		timestamps := timestampsStr == "true" || timestampsStr == "1"

		logs, err := h.service.GetPodLogs(c.Context(), namespace, podName, container, tailLines, timestamps)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, logs, "Pod logs retrieved successfully")
	}
}

// ListServices handles GET /api/v1/k8s/services.
func (h *K8sHandler) ListServices() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "dev-coffe")
		services, err := h.service.ListServices(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, services, "Services retrieved successfully")
	}
}

// GetService handles GET /api/v1/k8s/services/:namespace/:name.
func (h *K8sHandler) GetService() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		svc, err := h.service.GetService(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, svc, "Service details retrieved successfully")
	}
}

// ListIngresses handles GET /api/v1/k8s/ingresses.
func (h *K8sHandler) ListIngresses() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "dev-coffe")
		ingresses, err := h.service.ListIngresses(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, ingresses, "Ingresses retrieved successfully")
	}
}

// GetIngress handles GET /api/v1/k8s/ingresses/:namespace/:name.
func (h *K8sHandler) GetIngress() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		ing, err := h.service.GetIngress(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, ing, "Ingress details retrieved successfully")
	}
}

// ListCronJobs handles GET /api/v1/k8s/cronjobs.
func (h *K8sHandler) ListCronJobs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "dev-coffe")
		cronjobs, err := h.service.ListCronJobs(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, cronjobs, "CronJobs retrieved successfully")
	}
}

// GetCronJob handles GET /api/v1/k8s/cronjobs/:namespace/:name.
func (h *K8sHandler) GetCronJob() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		cj, err := h.service.GetCronJob(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, cj, "CronJob retrieved successfully")
	}
}

// CreateCronJob handles POST /api/v1/k8s/cronjobs.
func (h *K8sHandler) CreateCronJob() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateCronJobRequest
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequest("Invalid request payload")
		}
		if req.Name == "" || req.Namespace == "" || req.Schedule == "" {
			return errors.BadRequest("Name, Namespace, and Schedule are required")
		}

		cj, err := h.service.CreateCronJob(c.Context(), req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, cj, "CronJob created successfully")
	}
}

// UpdateCronJob handles PUT /api/v1/k8s/cronjobs/:namespace/:name.
func (h *K8sHandler) UpdateCronJob() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		var req UpdateCronJobRequest
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequest("Invalid request payload")
		}

		updated, err := h.service.UpdateCronJob(c.Context(), namespace, name, req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, updated, "CronJob updated successfully")
	}
}

// ToggleSuspendCronJob handles POST /api/v1/k8s/cronjobs/:namespace/:name/toggle-suspend.
func (h *K8sHandler) ToggleSuspendCronJob() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		suspended, err := h.service.ToggleSuspendCronJob(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"suspended": suspended}, "CronJob suspend toggled successfully")
	}
}

// TriggerCronJobNow handles POST /api/v1/k8s/cronjobs/:namespace/:name/run.
func (h *K8sHandler) TriggerCronJobNow() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		job, err := h.service.TriggerCronJobNow(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, job, "CronJob manual execution triggered successfully")
	}
}

// GetCronJobJobs handles GET /api/v1/k8s/cronjobs/:namespace/:name/jobs.
func (h *K8sHandler) GetCronJobJobs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		jobs, err := h.service.GetCronJobJobs(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, jobs, "CronJob execution jobs retrieved successfully")
	}
}

// DeleteCronJob handles DELETE /api/v1/k8s/cronjobs/:namespace/:name.
func (h *K8sHandler) DeleteCronJob() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		if err := h.service.DeleteCronJob(c.Context(), namespace, name); err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"deleted": true}, "CronJob deleted successfully")
	}
}

// ScaleDeployment handles PUT /api/v1/k8s/deployments/:namespace/:name/scale.
func (h *K8sHandler) ScaleDeployment() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		var req ScaleDeploymentRequest
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequest("Invalid request body")
		}
		if req.Replicas < 0 {
			return errors.BadRequest("Replicas cannot be negative")
		}
		dep, err := h.service.ScaleDeployment(c.Context(), namespace, name, req.Replicas)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, dep, "Deployment scaled successfully")
	}
}

// ListEvents handles GET /api/v1/k8s/events.
func (h *K8sHandler) ListEvents() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "dev-coffe")
		events, err := h.service.ListEvents(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, events, "Cluster events retrieved successfully")
	}
}

// ListPVCs handles GET /api/v1/k8s/pvcs.
func (h *K8sHandler) ListPVCs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace", "dev-coffe")
		pvcs, err := h.service.ListPVCs(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, pvcs, "PVCs retrieved successfully")
	}
}

// ListPVs handles GET /api/v1/k8s/pvs.
func (h *K8sHandler) ListPVs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		pvs, err := h.service.ListPVs(c.Context())
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, pvs, "PVs retrieved successfully")
	}
}

// ApplyYAML handles POST /api/v1/k8s/apply-yaml.
func (h *K8sHandler) ApplyYAML() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ApplyYAMLRequest
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequest("Invalid request body")
		}

		if req.YAML == "" {
			return errors.BadRequest("YAML manifest cannot be empty")
		}

		res, err := h.service.ApplyYAML(c.Context(), req.YAML, req.Namespace, req.DryRun)
		if err != nil {
			return err
		}

		msg := "YAML applied successfully"
		if req.DryRun {
			msg = "YAML validated successfully (dry run)"
		}

		return response.SuccessResponse(c, res, msg)
	}
}

// GetClusterOverview handles GET /api/v1/k8s/cluster-overview.
func (h *K8sHandler) GetClusterOverview() fiber.Handler {
	return func(c *fiber.Ctx) error {
		overview, err := h.service.GetClusterOverview(c.Context())
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, overview, "Cluster overview retrieved successfully")
	}
}

// ListNodes handles GET /api/v1/k8s/nodes.
func (h *K8sHandler) ListNodes() fiber.Handler {
	return func(c *fiber.Ctx) error {
		nodes, err := h.service.ListNodes(c.Context())
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, nodes, "Cluster nodes retrieved successfully")
	}
}

// ListPods handles GET /api/v1/k8s/pods.
func (h *K8sHandler) ListPods() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace")
		pods, err := h.service.ListPods(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, pods, "Pods retrieved successfully")
	}
}

// DeletePod handles DELETE /api/v1/k8s/pods/:namespace/:name.
func (h *K8sHandler) DeletePod() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		err := h.service.DeletePod(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"deleted": true, "name": name, "namespace": namespace}, "Pod deleted successfully")
	}
}

// ListStatefulSets handles GET /api/v1/k8s/statefulsets.
func (h *K8sHandler) ListStatefulSets() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace")
		ss, err := h.service.ListStatefulSets(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, ss, "StatefulSets retrieved successfully")
	}
}

// ScaleStatefulSet handles PUT /api/v1/k8s/statefulsets/:namespace/:name/scale.
func (h *K8sHandler) ScaleStatefulSet() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		var req ScaleStatefulSetRequest
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequest("Invalid request body")
		}
		if req.Replicas < 0 {
			return errors.BadRequest("Replicas cannot be negative")
		}

		err := h.service.ScaleStatefulSet(c.Context(), namespace, name, req.Replicas)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"scaled": true, "name": name, "namespace": namespace, "replicas": req.Replicas}, "StatefulSet scaled successfully")
	}
}

// RolloutRestartStatefulSet handles POST /api/v1/k8s/statefulsets/:namespace/:name/restart.
func (h *K8sHandler) RolloutRestartStatefulSet() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		err := h.service.RolloutRestartStatefulSet(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"restarted": true, "name": name, "namespace": namespace}, "StatefulSet rollout restart triggered successfully")
	}
}

// ListDaemonSets handles GET /api/v1/k8s/daemonsets.
func (h *K8sHandler) ListDaemonSets() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace")
		ds, err := h.service.ListDaemonSets(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, ds, "DaemonSets retrieved successfully")
	}
}

// RolloutRestartDaemonSet handles POST /api/v1/k8s/daemonsets/:namespace/:name/restart.
func (h *K8sHandler) RolloutRestartDaemonSet() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		err := h.service.RolloutRestartDaemonSet(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"restarted": true, "name": name, "namespace": namespace}, "DaemonSet rollout restart triggered successfully")
	}
}

// GetResourceYAML handles GET /api/v1/k8s/resource-yaml.
func (h *K8sHandler) GetResourceYAML() fiber.Handler {
	return func(c *fiber.Ctx) error {
		kind := c.Query("kind")
		namespace := c.Query("namespace")
		name := c.Query("name")

		if kind == "" || name == "" {
			return errors.BadRequest("Missing required query parameters: 'kind' and 'name'")
		}

		res, err := h.service.GetResourceYAML(c.Context(), kind, namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, res, "Resource YAML retrieved successfully")
	}
}

// ExecContainerTerminal handles WebSocket /api/v1/k8s/ws/exec/:namespace/:pod.
func (h *K8sHandler) ExecContainerTerminal() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		namespace := c.Params("namespace")
		pod := c.Params("pod")
		container := c.Query("container")
		shell := c.Query("shell", "sh")

		_ = h.service.ExecPodTerminal(context.Background(), c, namespace, pod, container, shell)
	})
}

// GetServiceEndpoints handles GET /api/v1/k8s/services/:namespace/:name/endpoints.
func (h *K8sHandler) GetServiceEndpoints() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Params("namespace")
		name := c.Params("name")
		ep, err := h.service.GetServiceEndpoints(c.Context(), namespace, name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, ep, "Service endpoints retrieved successfully")
	}
}

// GetPodMetrics handles GET /api/v1/k8s/metrics/pods.
func (h *K8sHandler) GetPodMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace")
		metrics, err := h.service.GetPodMetrics(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, metrics, "Pod metrics retrieved successfully")
	}
}

// CreateNamespace handles POST /api/v1/k8s/namespaces.
func (h *K8sHandler) CreateNamespace() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateNamespaceRequest
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequest("Invalid request body")
		}
		ns, err := h.service.CreateNamespace(c.Context(), &req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, ns, "Namespace created successfully")
	}
}

// DeleteNamespace handles DELETE /api/v1/k8s/namespaces/:name.
func (h *K8sHandler) DeleteNamespace() fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("name")
		err := h.service.DeleteNamespace(c.Context(), name)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, fiber.Map{"deleted": true, "name": name}, "Namespace deleted successfully")
	}
}

// GetResourceQuotas handles GET /api/v1/k8s/resource-quotas.
func (h *K8sHandler) GetResourceQuotas() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace")
		quotas, err := h.service.GetResourceQuotas(c.Context(), namespace)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, quotas, "Resource quotas retrieved successfully")
	}
}

// ListClusterEvents handles GET /api/v1/k8s/events/feed.
func (h *K8sHandler) ListClusterEvents() fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespace := c.Query("namespace")
		eventType := c.Query("type")
		kind := c.Query("kind")
		events, err := h.service.ListClusterEvents(c.Context(), namespace, eventType, kind)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, events, "Cluster events retrieved successfully")
	}
}



