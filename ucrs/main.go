package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/router"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/seeder"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/service"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// @title Underleaf Capability Registry API
// @version 1.0
// @description Global semantic authority and provider index for Underleaf Agents
// @description
// @description The Capability Registry Service (UCRS) provides:
// @description - Capability schema governance
// @description - Provider metadata registry
// @description - Trust tier classification
// @description - Artifact location indexing
// @description - Signed registry snapshot distribution

// @contact.name Ambient Labs Computing
// @contact.email support@ambientlabs.io

// @license.name Proprietary
// @license.url https://ambientlabs.io/license

// @host localhost:8083
// @BasePath /api/v1/registry

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"

func main() {
	ctx := context.Background()

	// Load settings
	settings := utils.LoadSettings()

	// Initialize logger
	logger, ctx := utils.InitLoggerWithContext(ctx, settings)
	logger.Info("starting Underleaf Capability Registry Service", "version", "1.0.0")

	// Build MongoDB client
	mongoClient, err := buildMongoClient(ctx, settings)
	if err != nil {
		logger.Error("failed to build MongoDB client", "error", err)
		panic(err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("failed to disconnect MongoDB", "error", err)
		}
	}()

	// Initialize repository
	repo := repository.NewMongoRepository(mongoClient.Database(settings.Mongo.MongoDatabase))
	logger.Info("initialized MongoDB repository", "database", settings.Mongo.MongoDatabase)

	// Load seed data
	seedLoader := seeder.NewSeeder(repo, settings.Environment, settings.Seeds.Path)
	if err := seedLoader.LoadSeeds(ctx); err != nil {
		logger.Error("failed to load seeds", "error", err)
		panic(err)
	}

	// Initialize service
	svc, err := service.NewAppService(repo, settings)
	if err != nil {
		logger.Error("failed to create service", "error", err)
		panic(err)
	}

	// Start service
	if err := svc.Start(ctx); err != nil {
		logger.Error("failed to start service", "error", err)
		panic(err)
	}

	// Initialize router
	appRouter, err := router.NewAppRouter(ctx, svc, settings)
	if err != nil {
		logger.Error("failed to create router", "error", err)
		panic(err)
	}

	// Start HTTP server in goroutine
	addr := settings.Address + ":" + settings.Port
	logger.Info("starting HTTP server", "address", addr)

	go func() {
		if err := appRouter.Engine().Run(addr); err != nil {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("Capability Registry Service started successfully",
		"port", settings.Port,
		"base_path", settings.BasePath)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down Capability Registry Service...")

	// Graceful shutdown
	if err := svc.Stop(ctx); err != nil {
		logger.Error("error during shutdown", "error", err)
	}

	logger.Info("Capability Registry Service stopped")
}

// buildMongoClient creates and connects a MongoDB client
func buildMongoClient(ctx context.Context, settings *utils.Settings) (*mongo.Client, error) {
	logger := utils.GetLogger(ctx)
	mongoURI := settings.Mongo.MongoURI
	username := settings.Mongo.MongoUser
	password := settings.Mongo.MongoPassword

	clientOptions := options.Client().ApplyURI(mongoURI)
	creds := options.Credential{
		Username: username,
		Password: string(password),
	}
	clientOptions.SetAuth(creds)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("failed to connect to MongoDB", "error", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("failed to ping MongoDB", "error", err)
		return nil, err
	}
	logger.Info("successfully connected and pinged MongoDB")
	return client, nil
}
