// Package app wires the whole application together: configuration, logging,
// and optional integrations (Redis, RabbitMQ, realtime) controlled by feature
// flags from the environment.
package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"golang/cache"
	"golang/config"
	"golang/database"
	k8smodule "golang/internal/module/k8s"
	"golang/pkg/auth"
	"golang/pkg/logging"
	queuepkg "golang/pkg/queue"
	"golang/pkg/realtime"
	"golang/pkg/telemetry"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// Container holds all initialized dependencies and their optional features:
//   - Cache  is nil when the Redis feature is disabled.
//   - Broker is nil when the RabbitMQ feature is disabled.
//   - Hub    is only initialized when the Realtime feature is enabled.
type Container struct {
	Config *config.Config
	DB     *database.PostgresDB

	Cache  *cache.RedisCache // nil when Redis feature is disabled
	Broker queuepkg.Broker   // nil when RabbitMQ feature is disabled
	Hub    *realtime.Hub     // nil when Realtime feature is disabled

	JWTService *auth.JWTService
	Logger     *slog.Logger
	K8sService *k8smodule.K8sService
}

// Boot loads the configuration, sets up the logger and every enabled
// integration, then returns the assembled container with a shutdown function
// that closes everything in reverse order.
func Boot() (*Container, func() error) {
	cfg := config.Load()
	logger := logging.Setup(cfg.Log.Level, cfg.Log.File)

	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Initialize OpenTelemetry
	if tp, err := telemetry.InitTracer(cfg); err != nil {
		logger.Warn("failed to initialize OpenTelemetry", logging.Err(err))
	} else if tp != nil {
		logger.Info("opentelemetry initialized")
	}

	logger.Info("configuration loaded",
		slog.String("env", cfg.App.Environment),
		slog.Group("features", slog.Bool("redis", cfg.Feature.Redis),
			slog.Bool("rabbitmq", cfg.Feature.RabbitMQ),
			slog.Bool("realtime", cfg.Feature.Realtime),
			slog.Bool("realtime_sse", cfg.Feature.RealtimeSSE),
			slog.Bool("realtime_ws", cfg.Feature.RealtimeWS),
			slog.String("realtime_bridge", cfg.Feature.RealtimeBridge)),
	)

	// Shutdown stack, closed in reverse (LIFO).
	var closers []func() error
	shutdown := func() error {
		logger.Info("shutting down services")
		var errs error
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i](); err != nil {
				logger.Error("failed to close resource", logging.Err(err))
				errs = err
			}
		}
		logging.Destroy()
		return errs
	}

	// Database connection (graceful fallback if offline)
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Warn("postgres unavailable, continuing with demo/k8s mode",
			slog.String("host", cfg.Database.Host),
			logging.Err(err),
		)
	} else {
		closers = append(closers, func() error { return db.Close() })
		container.DB = db
		logger.Info("postgres initialized")
		seedAdminUser(db.GetDB(), logger)
	}

	// Kubernetes service manager
	clientMgr := k8smodule.NewClientManager(logger)
	container.K8sService = k8smodule.NewK8sService(clientMgr)

	// Redis turns on JWT revocation tracking and distributed realtime
	// pub/sub. When disabled, those features degrade gracefully.
	if cfg.Feature.Redis {
		redisCache, err := cache.NewRedisCache(&cfg.Redis)
		if err != nil {
			logger.Warn("redis feature enabled but connection failed", logging.Err(err))
		} else {
			container.Cache = redisCache
			closers = append(closers, func() error { return redisCache.Close() })
			logger.Info("redis initialized")
		}
	} else {
		logger.Info("redis feature disabled")
	}

	// RabbitMQ broker. Skipped entirely when the feature is turned off.
	if cfg.Feature.RabbitMQ {
		broker, err := queuepkg.NewRabbitMQ(&cfg.RabbitMQ, logger)
		if err != nil {
			logger.Warn("rabbitmq enabled but connection failed; skipping", logging.Err(err))
		} else {
			container.Broker = broker
			closers = append(closers, func() error { return broker.Close() })
			logger.Info("rabbitmq initialized")
		}
	} else {
		logger.Info("rabbitmq feature disabled")
	}

	// JWT service; Redis is passed through so revocation works when enabled.
	jwtService := auth.NewJWTService(
		&auth.JWTConfig{
			AccessSecret:         cfg.JWT.AccessSecret,
			RefreshSecret:        cfg.JWT.RefreshSecret,
			AccessTokenDuration:  time.Duration(cfg.JWT.AccessTokenDuration) * time.Minute,
			RefreshTokenDuration: time.Duration(cfg.JWT.RefreshTokenDuration) * time.Hour,
			Issuer:               cfg.JWT.Issuer,
		},
		container.Cache,
	)
	container.JWTService = jwtService

	// Realtime (WebSocket and/or SSE). SSE and WS are activated through their
	// own flags; the hub fans out across instances using the configured
	// bridge (Redis, RabbitMQ or both). When no extra service is available the
	// hub still works in single-instance (local-only) mode.
	if cfg.Feature.RealtimeSSE || cfg.Feature.RealtimeWS {
		bridges, names := buildRealtimeBridges(cfg, container, logger)

		hub := realtime.NewHub(logger, bridges...)
		go hub.Run()
		closers = append(closers, func() error { hub.Shutdown(); return nil })
		container.Hub = hub
		if container.K8sService != nil {
			container.K8sService.SetHub(hub)
			container.K8sService.StartWatchers(context.Background())
		}
		logger.Info("realtime hub initialized and connected to kubernetes",
			slog.String("bridge", cfg.Feature.RealtimeBridge),
			slog.Any("bridges", names),
		)
	} else {
		logger.Info("realtime feature disabled")
	}

	return container, shutdown
}

// buildRealtimeBridges assembles the distributed fan-out transports requested
// by REALTIME_BRIDGE. A requested transport whose backing service is missing
// is skipped with a warning instead of failing the whole boot.
func buildRealtimeBridges(cfg *config.Config, container *Container, logger *slog.Logger) ([]realtime.Bridge, []string) {
	var (
		bridges []realtime.Bridge
		names   []string
	)
	useRedis := cfg.Feature.RealtimeBridge == config.BridgeRedis || cfg.Feature.RealtimeBridge == config.BridgeBoth
	useRabbit := cfg.Feature.RealtimeBridge == config.BridgeRabbit || cfg.Feature.RealtimeBridge == config.BridgeBoth

	if useRedis && container.Cache == nil {
		logger.Warn("realtime bridge set to redis/both but Redis is disabled; redis fan-out skipped")
		useRedis = false
	}
	if useRabbit && container.Broker == nil {
		logger.Warn("realtime bridge set to rabbitmq/both but RabbitMQ is disabled; rabbitmq fan-out skipped")
		useRabbit = false
	}

	if useRedis {
		if b, err := realtime.NewRedisBridge(container.Cache, logger); err != nil {
			logger.Warn("failed to init redis bridge", logging.Err(err))
		} else {
			bridges = append(bridges, b)
			names = append(names, b.Name())
		}
	}
	if useRabbit {
		if b, err := realtime.NewRabbitMQBridge(container.Broker, logger); err != nil {
			logger.Warn("failed to init rabbitmq bridge", logging.Err(err))
		} else {
			bridges = append(bridges, b)
			names = append(names, b.Name())
		}
	}
	return bridges, names
}

// seedAdminUser seeds or updates the primary cluster administrator account
// using the auto-generated or configured ADMIN_PASSWORD from Kubernetes Secret / environment.
func seedAdminUser(db *sqlx.DB, log *slog.Logger) {
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@kubenexus.local"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash seed admin password", logging.Err(err))
		return
	}

	query := `
		INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at)
		VALUES ('a0000000-0000-0000-0000-000000000001', 'Cluster Administrator', $1, $2, 'admin', true, NOW(), NOW())
		ON CONFLICT (email) DO UPDATE SET password = EXCLUDED.password, role = 'admin', updated_at = NOW()
	`
	if _, err := db.Exec(query, adminEmail, string(hashedPassword)); err != nil {
		log.Warn("seed admin user skipped or table not yet migrated", logging.Err(err))
		return
	}

	log.Info("seed admin account verified from environment/secret", slog.String("email", adminEmail))
}
