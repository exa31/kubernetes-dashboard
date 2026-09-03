// Package module contains the feature modules of the application. Every
// module is self-contained: it owns its DTOs, repository, service and
// handler, and registers its own routes.
package module

import (
	"golang/config"
	"golang/database"
	authmodule "golang/internal/module/auth"
	k8smodule "golang/internal/module/k8s"
	usermodule "golang/internal/module/user"
	"golang/pkg/auth"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/realtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Router bundles module dependencies and registers routes under a base group.
type Router struct {
	DB         *database.PostgresDB
	JWTService *auth.JWTService
	Hub        *realtime.Hub // nil when the realtime feature is disabled
	Config     *config.Config
	K8sService *k8smodule.K8sService
}

// NewRouter builds the module router.
func NewRouter(db *database.PostgresDB, jwtService *auth.JWTService, hub *realtime.Hub, cfg *config.Config, k8sService *k8smodule.K8sService) *Router {
	return &Router{DB: db, JWTService: jwtService, Hub: hub, Config: cfg, K8sService: k8sService}
}

// Register mounts every enabled module on the API router.
func (r *Router) Register(api fiber.Router) {
	r.registerAuth(api)
	r.registerUsers(api)
	r.registerProtected(api)
	r.registerK8s(api)
	if r.Hub != nil {
		r.registerRealtime(api)
	}
}

func (r *Router) registerAuth(api fiber.Router) {
	handler := authmodule.NewAuthHandler(authmodule.NewAuthService(r.DB, r.JWTService))

	auth := api.Group("/auth")
	auth.Post("/register", handler.Register())
	auth.Post("/login", handler.Login())
	auth.Post("/refresh", handler.RefreshToken())

	auth.Use(authMiddleware.AuthMiddleware(r.JWTService))
	auth.Get("/profile", handler.GetProfile())
	auth.Put("/profile", handler.UpdateProfile())
	auth.Post("/change-password", handler.ChangePassword())
	auth.Post("/logout", handler.Logout())
	auth.Post("/logout-all", handler.LogoutAll())
}

func (r *Router) registerUsers(api fiber.Router) {
	if r.DB == nil {
		return
	}
	repo := usermodule.NewUserRepository(r.DB.GetDB())
	service := usermodule.NewUserService(repo)
	handler := usermodule.NewUserHandler(service)

	users := api.Group("/users")
	users.Use(authMiddleware.AuthMiddleware(r.JWTService))

	// Authenticated users can list and get user details
	users.Get("/", handler.GetUsers())
	users.Get("/:id", handler.GetUser())

	// Management operations restricted to admin
	adminUsers := users.Group("", authMiddleware.RequireRole("admin"))
	adminUsers.Post("/", handler.CreateUser())
	adminUsers.Put("/:id", handler.UpdateUser())
	adminUsers.Post("/:id/reset-password", handler.ResetPassword())
	adminUsers.Delete("/:id", handler.DeleteUser())
	adminUsers.Delete("/admin/:id", handler.HardDeleteUser())
}

func (r *Router) registerProtected(api fiber.Router) {
	protected := api.Group("/protected")
	protected.Use(authMiddleware.AuthMiddleware(r.JWTService))
	protected.Get("/data", func(c *fiber.Ctx) error {
		userID, _ := authMiddleware.GetUserID(c)
		email, _ := authMiddleware.GetEmail(c)
		return c.JSON(fiber.Map{
			"message": "This is protected data",
			"user_id": userID,
			"email":   email,
		})
	})

	api.Get("/public-or-private", authMiddleware.OptionalAuthMiddleware(r.JWTService), func(c *fiber.Ctx) error {
		if userID, err := authMiddleware.GetUserID(c); err == nil {
			return c.JSON(fiber.Map{"message": "Private data", "authenticated": true, "user_id": userID})
		}
		return c.JSON(fiber.Map{"message": "Public data", "authenticated": false})
	})
}

func (r *Router) registerRealtime(api fiber.Router) {
	rt := api.Group("/realtime")

	if r.Config.Feature.RealtimeWS {
		r.registerWebSocketRoutes(rt)
	}
	if r.Config.Feature.RealtimeSSE {
		r.registerSSERoutes(rt)
	}
}

func (r *Router) registerWebSocketRoutes(rt fiber.Router) {
	wsHandler := NewWSHandler(r.Hub)
	ws := rt.Group("/ws")
	ws.Get("/connect", upgradeWS(), websocket.New(wsHandler.Handle))
	ws.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), wsHandler.GetStats)
}

func (r *Router) registerSSERoutes(rt fiber.Router) {
	sseHandler := NewSSEHandler(r.Hub)
	sse := rt.Group("/sse")
	sse.Get("/events", sseHandler.Events)
	sse.Get("/subscribe", sseHandler.Subscribe)
	sse.Post("/broadcast", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.BroadcastToChannel)
	sse.Post("/send", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.SendToUser)
	sse.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.GetStats)
}

func (r *Router) registerK8s(api fiber.Router) {
	if r.K8sService == nil {
		return
	}
	handler := k8smodule.NewK8sHandler(r.K8sService)
	k8s := api.Group("/k8s")
	k8s.Use(authMiddleware.AuthMiddleware(r.JWTService))

	k8s.Get("/cluster-info", handler.GetClusterInfo())
	k8s.Get("/namespaces", handler.ListNamespaces())

	k8s.Get("/secrets", handler.ListSecrets())
	k8s.Get("/secrets/:namespace/:name", handler.GetSecret())
	k8s.Post("/secrets", handler.SaveSecret())
	k8s.Delete("/secrets/:namespace/:name", handler.DeleteSecret())

	k8s.Get("/configmaps", handler.ListConfigMaps())
	k8s.Get("/configmaps/:namespace/:name", handler.GetConfigMap())
	k8s.Post("/configmaps", handler.SaveConfigMap())
	k8s.Delete("/configmaps/:namespace/:name", handler.DeleteConfigMap())

	k8s.Get("/deployments", handler.ListDeployments())
	k8s.Get("/deployments/:namespace/:name", handler.GetDeployment())
	k8s.Put("/deployments/:namespace/:name", handler.UpdateDeployment())
	k8s.Put("/deployments/:namespace/:name/scale", handler.ScaleDeployment())
	k8s.Post("/deployments/:namespace/:name/restart", handler.RolloutRestartDeployment())
	k8s.Get("/deployments/:namespace/:name/pods", handler.GetDeploymentPods())

	k8s.Get("/services", handler.ListServices())
	k8s.Get("/services/:namespace/:name", handler.GetService())

	k8s.Get("/ingresses", handler.ListIngresses())
	k8s.Get("/ingresses/:namespace/:name", handler.GetIngress())

	k8s.Get("/cronjobs", handler.ListCronJobs())
	k8s.Post("/cronjobs", handler.CreateCronJob())
	k8s.Get("/cronjobs/:namespace/:name", handler.GetCronJob())
	k8s.Put("/cronjobs/:namespace/:name", handler.UpdateCronJob())
	k8s.Post("/cronjobs/:namespace/:name/toggle-suspend", handler.ToggleSuspendCronJob())
	k8s.Post("/cronjobs/:namespace/:name/run", handler.TriggerCronJobNow())
	k8s.Get("/cronjobs/:namespace/:name/jobs", handler.GetCronJobJobs())
	k8s.Delete("/cronjobs/:namespace/:name", handler.DeleteCronJob())

	k8s.Get("/events", handler.ListEvents())
	k8s.Get("/pvcs", handler.ListPVCs())
	k8s.Get("/pvs", handler.ListPVs())

	k8s.Get("/pods/:namespace/:name/logs", handler.GetPodLogs())

	// Dynamic Resource Manifest Apply
	k8s.Post("/apply-yaml", handler.ApplyYAML())

	// Cluster Overview & Node Telemetry
	k8s.Get("/cluster-overview", handler.GetClusterOverview())
	k8s.Get("/nodes", handler.ListNodes())

	// Pod Management (Deep-Dive & Kill/Redeploy)
	k8s.Get("/pods", handler.ListPods())
	k8s.Delete("/pods/:namespace/:name", handler.DeletePod())

	// StatefulSets & DaemonSets Workloads
	k8s.Get("/statefulsets", handler.ListStatefulSets())
	k8s.Put("/statefulsets/:namespace/:name/scale", handler.ScaleStatefulSet())
	k8s.Post("/statefulsets/:namespace/:name/restart", handler.RolloutRestartStatefulSet())

	k8s.Get("/daemonsets", handler.ListDaemonSets())
	k8s.Post("/daemonsets/:namespace/:name/restart", handler.RolloutRestartDaemonSet())

	// In-Place Live Resource YAML Inspector
	k8s.Get("/resource-yaml", handler.GetResourceYAML())

	// Enterprise Extensions: Endpoints, Metrics, Namespaces, Quotas, Events Feed
	k8s.Get("/services/:namespace/:name/endpoints", handler.GetServiceEndpoints())
	k8s.Get("/metrics/pods", handler.GetPodMetrics())
	k8s.Post("/namespaces", handler.CreateNamespace())
	k8s.Delete("/namespaces/:name", handler.DeleteNamespace())
	k8s.Get("/resource-quotas", handler.GetResourceQuotas())
	k8s.Get("/events/feed", handler.ListClusterEvents())

	// Interactive Container Web Terminal WebSocket
	k8s.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	k8s.Get("/ws/exec/:namespace/:pod", handler.ExecContainerTerminal())
}

