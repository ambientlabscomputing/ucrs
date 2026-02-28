package router

import (
	"context"

	"github.com/ambientlabscomputing/ucrs/service"
	"github.com/ambientlabscomputing/ucrs/utils"
	"github.com/gin-gonic/gin"
)

type AppRouter struct {
	engine   *gin.Engine
	service  service.Service
	settings *utils.Settings
}

func NewAppRouter(ctx context.Context, svc service.Service, settings *utils.Settings) (*AppRouter, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Global middleware
	engine.Use(SlogRecoveryMiddleware(ctx))
	engine.Use(TraceIDMiddleware(ctx))
	engine.Use(SlogLoggerMiddleware(ctx))
	engine.Use(CORSMiddleware())

	router := &AppRouter{
		engine:   engine,
		service:  svc,
		settings: settings,
	}

	// Initialize JWKS
	if err := StartMiddleware(ctx, settings); err != nil {
		return nil, err
	}

	// Setup routes
	router.setupRoutes(ctx)

	return router, nil
}

func (r *AppRouter) Engine() *gin.Engine {
	return r.engine
}

func (r *AppRouter) setupRoutes(ctx context.Context) {
	basePath := r.settings.BasePath

	// Health check (no auth)
	r.engine.GET(basePath+"/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public key endpoint (no auth) - for signature verification
	r.engine.GET(basePath+"/public-key", r.GetPublicKeyHandler())

	// Sync endpoints (read-only, PUBLIC - no auth required)
	syncGroup := r.engine.Group(basePath + "/sync")
	{
		syncGroup.GET("/snapshot", r.GetSnapshotHandler())
		syncGroup.GET("/delta", r.GetDeltaHandler())
	}

	// Capability endpoints
	capabilitiesGroup := r.engine.Group(basePath + "/capabilities")
	{
		// Read operations (PUBLIC - no auth)
		capabilitiesGroup.GET("", r.ListCapabilitiesHandler())
		capabilitiesGroup.GET("/:id", r.GetCapabilityHandler())

		// Write operations (write:registry scope)
		capabilitiesGroup.POST("", WriteAuthMiddleware(ctx, r.settings), r.CreateCapabilityHandler())
		capabilitiesGroup.PUT("/:id", WriteAuthMiddleware(ctx, r.settings), r.UpdateCapabilityHandler())
		capabilitiesGroup.DELETE("/:id", AdminAuthMiddleware(ctx, r.settings), r.DeleteCapabilityHandler())
	}

	// Provider endpoints
	providersGroup := r.engine.Group(basePath + "/providers")
	{
		// Read operations (PUBLIC - no auth)
		providersGroup.GET("", r.ListProvidersHandler())
		providersGroup.GET("/:provider_id/:version", r.GetProviderHandler())

		// Write operations (write:registry scope)
		providersGroup.POST("", WriteAuthMiddleware(ctx, r.settings), r.CreateProviderHandler())
		providersGroup.PUT("/:provider_id/:version", WriteAuthMiddleware(ctx, r.settings), r.UpdateProviderHandler())
		providersGroup.DELETE("/:provider_id/:version", AdminAuthMiddleware(ctx, r.settings), r.DeleteProviderHandler())
	}
}
