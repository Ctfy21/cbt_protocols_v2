package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"backend_v2/internal/config"
	"backend_v2/internal/database"
	"backend_v2/internal/handlers"
	"backend_v2/internal/middleware"
	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to MongoDB
	db, err := database.Connect(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	log.Println("✅ Connected to MongoDB")

	chamberService := services.NewChamberService(db, cfg)
	experimentService := services.NewExperimentService(db)
	authService := services.NewAuthService(db, cfg)
	apiTokenService := services.NewAPITokenService(db)
	userChamberAccessService := services.NewUserChamberAccessService(db)

	// Initialize handlers
	chamberHandler := handlers.NewChamberHandler(chamberService)
	experimentHandler := handlers.NewExperimentHandler(experimentService)
	authHandler := handlers.NewAuthHandler(authService)
	apiTokenHandler := handlers.NewAPITokenHandler(apiTokenService)
	userChamberAccessHandler := handlers.NewUserChamberAccessHandler(userChamberAccessService)
	userHandler := handlers.NewUserManagementHandler(authService)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Create router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup API routes
	setupAPIRoutes(router, chamberHandler, experimentHandler, authHandler, apiTokenHandler, userChamberAccessHandler, userHandler, apiTokenService, authService)

	// Setup frontend routes
	setupFrontendRoutes(router)

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go chamberService.StartStatusMonitor(ctx)
	log.Println("✅ Started chamber status monitor")

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("🌐 Frontend available at: http://localhost:%s", cfg.Port)
		log.Printf("🔧 API available at: http://localhost:%s/api", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏸️  Shutting down server...")

	// Cancel background services
	cancel()

	// Shutdown server with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server shutdown complete")
}

func setupAPIRoutes(
	router *gin.Engine,
	chamberHandler *handlers.ChamberHandler,
	experimentHandler *handlers.ExperimentHandler,
	authHandler *handlers.AuthHandler,
	apiTokenHandler *handlers.APITokenHandler,
	userChamberAccessHandler *handlers.UserChamberAccessHandler,
	userHandler *handlers.UserManagementHandler,
	apiTokenService *services.APITokenService,
	authService *services.AuthService,
) {
	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
			"status": "healthy",
			"time":   time.Now(),
		}))
	})

	// API routes
	api := router.Group("/api")

	// Public auth routes
	api.POST("/auth/login", authHandler.Login)

	api.Use(middleware.AuthMiddleware(authService, apiTokenService))
	{

		// Auth routes
		api.POST("/auth/refresh", authHandler.RefreshToken)
		api.POST("/auth/logout", authHandler.Logout)
		api.GET("/auth/me", authHandler.Me)
		api.PUT("/auth/profile", authHandler.UpdateProfile)
		api.POST("/auth/change-password", authHandler.ChangePassword)

		// API Token routes
		api.POST("/api-tokens", apiTokenHandler.CreateAPIToken)
		api.GET("/api-tokens", apiTokenHandler.GetAPITokens)
		api.DELETE("/api-tokens/:id", apiTokenHandler.RevokeAPIToken)

		// Chamber routes
		api.POST("/chambers", chamberHandler.RegisterChamber)
		api.POST("/chambers/:id/heartbeat", chamberHandler.Heartbeat)
		api.GET("/chambers/:id", chamberHandler.GetChamber)
		api.GET("/chambers", chamberHandler.GetChambers)
		api.GET("/chambers/:id/watering-zones", chamberHandler.GetChamberWateringZones)
		api.PUT("/chambers/:id/config", chamberHandler.UpdateChamberConfig)
		api.GET("/chambers/:id/config", chamberHandler.GetChamberConfig)
		api.GET("/chambers/:id/config/check", chamberHandler.CheckChamberConfigUpdate)

		// Experiment routes
		api.GET("/experiments/:id", experimentHandler.GetExperiment)
		api.GET("/experiments", experimentHandler.GetExperiments)
		api.POST("/experiments", experimentHandler.CreateExperiment)
		api.PUT("/experiments/:id", experimentHandler.UpdateExperiment)
		api.PATCH("/experiments/:id/status", experimentHandler.UpdateExperimentStatus)
		api.DELETE("/experiments/:id", experimentHandler.DeleteExperiment)

		// User Chamber Access routes (Admin only)
		adminRoutes := api.Group("/")
		adminRoutes.Use(middleware.RequireRole(models.RoleAdmin))
		{
			// User management routes
			adminRoutes.POST("/users", userHandler.CreateUser)
			adminRoutes.GET("/users", userHandler.GetUsers)
			adminRoutes.GET("/users/:id", userHandler.GetUser)
			adminRoutes.PUT("/users/:id", userHandler.UpdateUser)
			adminRoutes.DELETE("/users/:id", userHandler.DeactivateUser)
			adminRoutes.POST("/users/:id/activate", userHandler.ActivateUser)

			// User chamber access management
			adminRoutes.GET("/users/chambers", userChamberAccessHandler.GetAllUsersWithChamberAccess)
			adminRoutes.PUT("/users/:id/chambers", userChamberAccessHandler.SetUserChamberAccess)
			adminRoutes.GET("/users/:id/chambers", userChamberAccessHandler.GetUserChamberAccess)
			adminRoutes.POST("/users/:id/chambers/:chamber_id", userChamberAccessHandler.GrantChamberAccess)
			adminRoutes.DELETE("/users/:id/chambers/:chamber_id", userChamberAccessHandler.RevokeChamberAccess)
			adminRoutes.GET("/users/:id/chambers/:chamber_id/check", userChamberAccessHandler.HasChamberAccess)

		}

		// User's own chamber access (non-admin users can check their own access)
		api.GET("/me/chambers", func(c *gin.Context) {
			// Get user from context
			userInterface, exists := c.Get("user")
			if !exists {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse("User not found"))
				return
			}

			user, ok := userInterface.(*models.User)
			if !ok {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse("Invalid user data"))
				return
			}

			// Use the user chamber access handler but with current user's ID
			c.Params = gin.Params{{Key: "id", Value: user.ID.Hex()}}
			userChamberAccessHandler.GetUserChamberAccess(c)
		})

		// User's own room chambers access
		api.GET("/me/room-chambers", func(c *gin.Context) {
			// Get user from context
			userInterface, exists := c.Get("user")
			if !exists {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse("User not found"))
				return
			}

			user, ok := userInterface.(*models.User)
			if !ok {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse("Invalid user data"))
				return
			}

			// Use the user chamber access handler but with current user's ID
			c.Params = gin.Params{{Key: "id", Value: user.ID.Hex()}}
			userChamberAccessHandler.GetUserChamberAccess(c)
		})
	}
}

func setupFrontendRoutes(router *gin.Engine) {
	// Get the frontend dist subdirectory from embedded FS
	frontendDistFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Printf("Warning: Could not load embedded frontend files: %v", err)
		log.Println("Frontend will not be served. Make sure to build frontend and place dist files in backend/frontend/dist/")
		return
	}

	// Get the assets subdirectory from the dist directory
	assetsFS, err := fs.Sub(frontendDistFS, "assets")
	if err != nil {
		log.Printf("Warning: Could not load frontend assets: %v", err)
		// Fallback to serving the entire dist directory
		router.StaticFS("/assets", http.FS(frontendDistFS))
	} else {
		// Serve static files from the assets subdirectory
		router.StaticFS("/assets", http.FS(assetsFS))
	}

	// Serve favicon and other root files
	router.GET("/favicon.ico", func(c *gin.Context) {
		content, err := frontendDistFS.Open("favicon.ico")
		if err != nil {
			c.Status(404)
			return
		}
		defer content.Close()

		stat, err := content.Stat()
		if err != nil {
			c.Status(404)
			return
		}

		c.DataFromReader(200, stat.Size(), "image/x-icon", content, nil)
	})

	// Serve vite.svg
	router.GET("/vite.svg", func(c *gin.Context) {
		content, err := frontendDistFS.Open("vite.svg")
		if err != nil {
			c.Status(404)
			return
		}
		defer content.Close()

		stat, err := content.Stat()
		if err != nil {
			c.Status(404)
			return
		}

		c.DataFromReader(200, stat.Size(), "image/svg+xml", content, nil)
	})

	// Serve index.html for all non-API routes (SPA routing)
	router.NoRoute(func(c *gin.Context) {
		// Skip API routes
		if gin.IsDebugging() {
			log.Printf("Serving frontend for route: %s", c.Request.URL.Path)
		}

		if c.Request.URL.Path != "/" &&
			!gin.IsDebugging() &&
			c.GetHeader("Accept") != "" &&
			c.GetHeader("Accept") != "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" {
			// If it's an API request that doesn't exist, return 404
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}

		indexHTML, err := frontendDistFS.Open("index.html")
		if err != nil {
			log.Printf("Error serving index.html: %v", err)
			c.String(500, "Frontend not available")
			return
		}
		defer indexHTML.Close()

		stat, err := indexHTML.Stat()
		if err != nil {
			c.String(500, "Error reading index.html")
			return
		}

		c.DataFromReader(200, stat.Size(), "text/html; charset=utf-8", indexHTML, nil)
	})
	log.Println("✅ Frontend routes configured")
}
