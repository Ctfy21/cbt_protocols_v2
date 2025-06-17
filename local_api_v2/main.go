package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"local_api_v2/internal/config"
	"local_api_v2/internal/database"
	"local_api_v2/internal/services"
	"local_api_v2/pkg/homeassistant"
	"local_api_v2/pkg/ntp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Local API v2 with chamber suffixes: %v", cfg.ChamberSuffixes)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize NTP service
	ntpService := ntp.NewTimeService(ntp.Config{
		Enabled:      cfg.NTPEnabled,
		Servers:      cfg.NTPServers,
		Timeout:      cfg.NTPTimeout,
		SyncInterval: cfg.NTPSyncInterval,
	})

	// Start NTP service
	if err := ntpService.Start(ctx, cfg.NTPSyncInterval); err != nil {
		log.Printf("Warning: Failed to start NTP service: %v", err)
	}

	// Initialize database
	db, err := database.NewMongoDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from database: %v", err)
		}
	}()

	// Initialize Home Assistant client
	haClient := homeassistant.NewClient(cfg.HomeAssistantURL, cfg.HomeAssistantToken)

	// Initialize services
	discoveryService := services.NewDiscoveryService(haClient)
	chamberManager := services.NewChamberManager(cfg, db, discoveryService, ntpService)
	registrationService := services.NewRegistrationService(cfg, db, ntpService)
	syncService := services.NewSyncService(cfg, db, ntpService)

	// Set cross-references
	syncService.SetChamberManager(chamberManager)
	syncService.SetRegistrationService(registrationService)

	// Use WaitGroups and channels for proper synchronization
	var (
		wg                    sync.WaitGroup
		chamberInitialized    = make(chan struct{})
		registrationCompleted = make(chan struct{})
		executorServices      []*services.ExecutorService
		mu                    sync.Mutex
	)

	// Step 1: Wait for Home Assistant connection and initialize chambers
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chamberInitialized)

		for {
			if haClient.IsConnected() {
				log.Printf("Home Assistant connected. Discovering chambers...")

				// Initialize chambers
				if err := chamberManager.InitializeChambers(ctx); err != nil {
					log.Printf("Warning: Chamber initialization failed: %v", err)
					time.Sleep(10 * time.Second)
					continue
				}

				// Log discovered chambers
				chambers := chamberManager.GetChambers()
				log.Printf("Successfully discovered %d chambers:", len(chambers))

				for suffix, chamber := range chambers {
					log.Printf("  Chamber '%s': %s (%d inputs, %d lamps, %d zones)",
						suffix, chamber.Name,
						len(chamber.Config.InputNumbers),
						len(chamber.Config.Lamps),
						len(chamber.Config.WateringZones))

					// Create executor service for each chamber
					executor := services.NewExecutorService(db, haClient, chamber.ID, ntpService)
					mu.Lock()
					executorServices = append(executorServices, executor)
					mu.Unlock()
				}

				haClient.Status = true
				break
			}
			log.Println("Waiting for Home Assistant connection...")
			time.Sleep(10 * time.Second)
		}
	}()

	// Step 2: Register chambers with backend
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(registrationCompleted)

		// Wait for chamber initialization
		<-chamberInitialized

		chambers := chamberManager.GetChambers()
		if len(chambers) == 0 {
			log.Printf("No chambers discovered")
			return
		}

		log.Println("Registering chambers with backend...")
		if err := chamberManager.RegisterChambersWithBackend(registrationService); err != nil {
			log.Printf("Warning: Chamber registration failed: %v", err)
			// Don't return - we can still function without backend registration
		}
	}()

	// Step 3: Start background services after registration
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait for registration to complete (or fail)
		<-registrationCompleted

		// Start heartbeat service
		go func() {
			log.Println("Starting heartbeat service...")
			registrationService.StartHeartbeat(ctx, chamberManager)
		}()

		// Start sync service
		go func() {
			log.Println("Starting sync service...")
			syncService.StartSync(ctx)
		}()

		// Start executor services for each chamber
		time.Sleep(2 * time.Second) // Give sync service time to fetch experiments

		mu.Lock()
		defer mu.Unlock()

		for i, executor := range executorServices {
			if executor != nil {
				log.Printf("Starting executor service %d...", i+1)
				if err := executor.Start(ctx); err != nil {
					log.Printf("Warning: Failed to start executor service %d: %v", i+1, err)
				}
			}
		}
	}()

	// Start simple HTTP server for health checks
	mux := http.NewServeMux()
	setupRoutes(mux, db, chamberManager, ntpService, syncService)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🚀 Local API v2 initialization complete - waiting for shutdown signal...")

	<-quit

	log.Println("Shutting down server...")

	// Stop NTP service
	ntpService.Stop()

	// Stop executor services
	mu.Lock()
	for i, executor := range executorServices {
		if executor != nil {
			log.Printf("Stopping executor service %d...", i+1)
			executor.Stop()
		}
	}
	mu.Unlock()

	// Cancel context to stop background services
	cancel()

	// Shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupRoutes configures HTTP routes
func setupRoutes(mux *http.ServeMux, db *database.MongoDB, chamberManager *services.ChamberManager, ntpService *ntp.TimeService, syncService *services.SyncService) {
	// Health check endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		chambers := chamberManager.GetChambers()
		registeredCount := 0
		for _, chamber := range chambers {
			if !chamber.BackendID.IsZero() {
				registeredCount++
			}
		}

		fmt.Fprintf(w, `{"status":"healthy","total_chambers":%d,"registered_chambers":%d,"ntp_enabled":%t,"ntp_connected":%t}`,
			len(chambers), registeredCount, ntpService.IsEnabled(), ntpService.IsConnected())
	})

	// Sync status endpoint
	mux.HandleFunc("/api/v1/sync/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		status := syncService.GetSyncStatus()
		statusJSON, _ := json.Marshal(status)
		w.Write(statusJSON)
	})

	// Chambers endpoint
	mux.HandleFunc("/api/v1/chambers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		chambers := chamberManager.GetChambers()
		var chamberList []map[string]interface{}

		for suffix, chamber := range chambers {
			chamberList = append(chamberList, map[string]interface{}{
				"id":          chamber.ID.Hex(),
				"name":        chamber.Name,
				"suffix":      suffix,
				"backend_id":  chamber.BackendID.Hex(),
				"status":      chamber.Status,
				"registered":  !chamber.BackendID.IsZero(),
				"input_count": len(chamber.Config.InputNumbers),
				"lamp_count":  len(chamber.Config.Lamps),
				"zone_count":  len(chamber.Config.WateringZones),
			})
		}

		response, _ := json.Marshal(map[string]interface{}{
			"success": true,
			"data":    chamberList,
		})
		w.Write(response)
	})

	// Time endpoint (returns current time from NTP or system)
	mux.HandleFunc("/api/v1/time", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		currentTime := ntpService.Now()
		moscowTime := ntpService.NowInMoscow()

		fmt.Fprintf(w, `{"current_time":"%s","moscow_time":"%s","unix_timestamp":%d,"ntp_enabled":%t,"ntp_connected":%t}`,
			currentTime.Format("2006-01-02T15:04:05Z07:00"),
			moscowTime.Format("2006-01-02T15:04:05Z07:00"),
			ntpService.Unix(),
			ntpService.IsEnabled(),
			ntpService.IsConnected())
	})
}
