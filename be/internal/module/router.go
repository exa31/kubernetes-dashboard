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

	// Cluster-level telemetry (open to all authenticated users)
	k8s.Get("/cluster-info", handler.GetClusterInfo())
	k8s.Get("/namespaces", handler.ListNamespaces())
	k8s.Get("/cluster-overview", handler.GetClusterOverview())
	k8s.Get("/nodes", handler.ListNodes())
	k8s.Get("/events", handler.ListEvents())
	k8s.Get("/events/feed", handler.ListClusterEvents())
	k8s.Get("/pvs", handler.ListPVs())

	// Namespace-scoped Read Operations (Read access check)
	nsRead := k8s.Group("", authMiddleware.RequireNamespaceAccess(false))
	nsRead.Get("/secrets", handler.ListSecrets())
	nsRead.Get("/secrets/:namespace/:name", handler.GetSecret())
	nsRead.Get("/configmaps", handler.ListConfigMaps())
	nsRead.Get("/configmaps/:namespace/:name", handler.GetConfigMap())
	nsRead.Get("/deployments", handler.ListDeployments())
	nsRead.Get("/deployments/:namespace/:name", handler.GetDeployment())
	nsRead.Get("/deployments/:namespace/:name/pods", handler.GetDeploymentPods())
	nsRead.Get("/services", handler.ListServices())
	nsRead.Get("/services/:namespace/:name", handler.GetService())
	nsRead.Get("/services/:namespace/:name/endpoints", handler.GetServiceEndpoints())
	nsRead.Get("/ingresses", handler.ListIngresses())
	nsRead.Get("/ingresses/:namespace/:name", handler.GetIngress())
	nsRead.Get("/cronjobs", handler.ListCronJobs())
	nsRead.Get("/cronjobs/:namespace/:name", handler.GetCronJob())
	nsRead.Get("/cronjobs/:namespace/:name/jobs", handler.GetCronJobJobs())
	nsRead.Get("/pvcs", handler.ListPVCs())
	nsRead.Get("/pods", handler.ListPods())
	nsRead.Get("/pods/:namespace/:name/logs", handler.GetPodLogs())
	nsRead.Get("/statefulsets", handler.ListStatefulSets())
	nsRead.Get("/daemonsets", handler.ListDaemonSets())
	nsRead.Get("/resource-yaml", handler.GetResourceYAML())
	nsRead.Get("/metrics/pods", handler.GetPodMetrics())
	nsRead.Get("/resource-quotas", handler.GetResourceQuotas())

	// DevOps & Admin Mutating Operations (RequireRole "admin", "devops" & Write namespace check)
	devops := k8s.Group("", authMiddleware.RequireRole("admin", "devops"), authMiddleware.RequireNamespaceAccess(true))
	devops.Post("/secrets", handler.SaveSecret())
	devops.Delete("/secrets/:namespace/:name", handler.DeleteSecret())
	devops.Post("/configmaps", handler.SaveConfigMap())
	devops.Delete("/configmaps/:namespace/:name", handler.DeleteConfigMap())
	devops.Put("/deployments/:namespace/:name", handler.UpdateDeployment())
	devops.Put("/deployments/:namespace/:name/scale", handler.ScaleDeployment())
	devops.Post("/deployments/:namespace/:name/restart", handler.RolloutRestartDeployment())
	devops.Post("/cronjobs", handler.CreateCronJob())
	devops.Put("/cronjobs/:namespace/:name", handler.UpdateCronJob())
	devops.Post("/cronjobs/:namespace/:name/toggle-suspend", handler.ToggleSuspendCronJob())
	devops.Post("/cronjobs/:namespace/:name/run", handler.TriggerCronJobNow())
	devops.Delete("/cronjobs/:namespace/:name", handler.DeleteCronJob())
	devops.Post("/apply-yaml", handler.ApplyYAML())
	devops.Delete("/pods/:namespace/:name", handler.DeletePod())
	devops.Put("/statefulsets/:namespace/:name/scale", handler.ScaleStatefulSet())
	devops.Post("/statefulsets/:namespace/:name/restart", handler.RolloutRestartStatefulSet())
	devops.Post("/daemonsets/:namespace/:name/restart", handler.RolloutRestartDaemonSet())

	// Admin-Only Cluster Management (RequireRole "admin")
	adminK8s := k8s.Group("", authMiddleware.RequireRole("admin"))
	adminK8s.Post("/namespaces", handler.CreateNamespace())
	adminK8s.Delete("/namespaces/:name", handler.DeleteNamespace())

	// Interactive Container Web Terminal WebSocket (DevOps & Admin with namespace write access)
	k8s.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	k8s.Get("/ws/exec/:namespace/:pod",
		authMiddleware.RequireRole("admin", "devops"),
		authMiddleware.RequireNamespaceAccess(true),
		handler.ExecContainerTerminal(),
	)
}

