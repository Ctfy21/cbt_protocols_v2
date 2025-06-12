package main

import (
	"context"
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
	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

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

	// Initialize services
	chamberService := services.NewChamberService(db, cfg)
	experimentService := services.NewExperimentService(db)

	// Initialize handlers
	chamberHandler := handlers.NewChamberHandler(chamberService)
	experimentHandler := handlers.NewExperimentHandler(experimentService)

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

	// Setup routes
	setupRoutes(router, chamberHandler, experimentHandler)

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

func setupRoutes(router *gin.Engine, chamberHandler *handlers.ChamberHandler, experimentHandler *handlers.ExperimentHandler) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
			"status": "healthy",
			"time":   time.Now(),
		}))
	})

	// API routes
	api := router.Group("/")
	{
		// Chamber routes
		api.POST("/chambers", chamberHandler.RegisterChamber)
		api.POST("/chambers/:id/heartbeat", chamberHandler.Heartbeat)
		api.GET("/chambers/:id", chamberHandler.GetChamber)
		api.GET("/chambers", chamberHandler.GetChambers)
		api.GET("/chambers/:id/watering-zones", chamberHandler.GetChamberWateringZones)

		// Experiment routes
		api.POST("/experiments", experimentHandler.CreateExperiment)
		api.GET("/experiments/:id", experimentHandler.GetExperiment)
		api.GET("/experiments", experimentHandler.GetExperiments)
		api.PUT("/experiments/:id", experimentHandler.UpdateExperiment)
		api.DELETE("/experiments/:id", experimentHandler.DeleteExperiment)
	}
}
