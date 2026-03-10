package main

import (
	"fmt"
	"log"
	"time"

	"apihub/internal/config"
	"apihub/internal/database"
	adminhandler "apihub/internal/handlers/admin"
	proxyhandler "apihub/internal/handlers/proxy"
	"apihub/internal/middleware"
	"apihub/internal/proxy"
	"apihub/internal/repository"
	"apihub/internal/services"
	"apihub/pkg"
	"apihub/web"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := database.Init(cfg)

	httpClient := pkg.NewHTTPClient(120 * time.Second)

	// Repositories
	adminRepo := repository.NewAdminRepo(db)
	providerRepo := repository.NewProviderRepo(db)
	modelConfigRepo := repository.NewModelConfigRepo(db)
	apiKeyRepo := repository.NewAPIKeyRepo(db)
	logRepo := repository.NewLogRepo(db)

	// Services
	authService := services.NewAuthService(adminRepo)
	providerService := services.NewProviderService(providerRepo)
	modelConfigService := services.NewModelConfigService(modelConfigRepo, providerRepo)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo, modelConfigRepo)
	syncService := services.NewSyncService(providerRepo, providerService, httpClient)

	// Proxy core
	lb := proxy.NewLoadBalancer()
	resolver := proxy.NewModelResolver(apiKeyRepo, modelConfigRepo, lb)
	openaiProxy := proxy.NewOpenAIProxy(resolver, httpClient, logRepo)
	anthropicProxy := proxy.NewAnthropicProxy(resolver, httpClient, logRepo)

	// Handlers
	authHandler := adminhandler.NewAuthHandler(authService)
	providerHandler := adminhandler.NewProviderHandler(providerService, syncService)
	modelConfigHandler := adminhandler.NewModelConfigHandler(modelConfigService)
	apiKeyHandler := adminhandler.NewAPIKeyHandler(apiKeyService)
	logHandler := adminhandler.NewLogHandler(logRepo)
	openaiHandler := proxyhandler.NewOpenAIHandler(openaiProxy)
	anthropicHandler := proxyhandler.NewAnthropicHandler(anthropicProxy)

	// 启动 per-provider 定时同步调度器
	syncService.StartPerProviderScheduler()
	defer syncService.Stop()

	// Router
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Admin routes
	adminGroup := r.Group("/admin")
	{
		adminGroup.POST("/login", authHandler.Login)

		authed := adminGroup.Group("")
		authed.Use(middleware.AdminAuth(db))
		{
			authed.GET("/dashboard", logHandler.Dashboard)

			authed.GET("/providers", providerHandler.List)
			authed.GET("/providers/tags", providerHandler.ListTags)
			authed.POST("/providers", providerHandler.Create)
			authed.POST("/providers/sync-all", providerHandler.SyncAll)
			authed.GET("/providers/:id", providerHandler.Get)
			authed.PUT("/providers/:id", providerHandler.Update)
			authed.DELETE("/providers/:id", providerHandler.Delete)
			authed.PUT("/providers/:id/toggle", providerHandler.Toggle)
			authed.POST("/providers/:id/sync", providerHandler.Sync)
			authed.GET("/providers/:id/models", providerHandler.ListModels)
			authed.PUT("/providers/:id/models/:mid/toggle", providerHandler.ToggleModel)

			// ModelConfig management
			authed.GET("/model-configs", modelConfigHandler.List)
			authed.POST("/model-configs", modelConfigHandler.Create)
			authed.GET("/model-configs/:id", modelConfigHandler.Get)
			authed.PUT("/model-configs/:id", modelConfigHandler.Update)
			authed.DELETE("/model-configs/:id", modelConfigHandler.Delete)
			authed.PUT("/model-configs/:id/toggle", modelConfigHandler.Toggle)
			authed.POST("/model-configs/:id/clone", modelConfigHandler.Clone)
			authed.GET("/model-configs/available-models", modelConfigHandler.GetAllAvailableModels)
			authed.GET("/model-configs/:id/items", modelConfigHandler.ListItems)
			authed.PUT("/model-configs/:id/items", modelConfigHandler.BatchUpdate)
			authed.POST("/model-configs/:id/items", modelConfigHandler.AddItem)
			authed.PUT("/model-configs/:id/items/:iid", modelConfigHandler.UpdateItem)
			authed.DELETE("/model-configs/:id/items/:iid", modelConfigHandler.DeleteItem)
			authed.GET("/model-configs/:id/items/grouped", modelConfigHandler.ListItemsGrouped)

			// API Key management
			authed.GET("/apikeys", apiKeyHandler.List)
			authed.POST("/apikeys", apiKeyHandler.Create)
			authed.GET("/apikeys/:id", apiKeyHandler.Get)
			authed.PUT("/apikeys/:id", apiKeyHandler.Update)
			authed.DELETE("/apikeys/:id", apiKeyHandler.Delete)
			authed.PUT("/apikeys/:id/toggle", apiKeyHandler.Toggle)
			authed.GET("/apikeys/:id/models", apiKeyHandler.ListModels)
			authed.PUT("/apikeys/:id/models", apiKeyHandler.BatchUpdate)
			authed.POST("/apikeys/:id/models", apiKeyHandler.AddModel)
			authed.PUT("/apikeys/:id/models/:mid", apiKeyHandler.UpdateModel)
			authed.DELETE("/apikeys/:id/models/:mid", apiKeyHandler.DeleteModel)
			authed.GET("/apikeys/:id/models/grouped", apiKeyHandler.ListModelsGrouped)
			authed.GET("/apikeys/:id/model-config", apiKeyHandler.GetModelConfig)
			authed.PUT("/apikeys/:id/model-config", apiKeyHandler.SetModelConfig)

			// Logs
			authed.GET("/logs", logHandler.List)
		}
	}

	// Proxy routes - OpenAI protocol
	v1 := r.Group("/v1")
	v1.Use(middleware.APIKeyAuth(db))
	{
		v1.POST("/chat/completions", openaiHandler.ChatCompletions)
		v1.POST("/completions", openaiHandler.Completions)
		v1.GET("/models", openaiHandler.ListModels)
	}

	// Proxy routes - Anthropic protocol
	anthropic := r.Group("/anthropic")
	anthropic.Use(middleware.APIKeyAuth(db))
	{
		anthropic.POST("/v1/messages", anthropicHandler.Messages)
		anthropic.GET("/v1/models", anthropicHandler.ListModels)
	}

	// Serve embedded frontend
	web.ServeStatic(r)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("ApiHub starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
