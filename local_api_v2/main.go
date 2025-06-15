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
	"local_api_v2/internal/models"
	"local_api_v2/internal/services"
	"local_api_v2/pkg/homeassistant"
	"local_api_v2/pkg/ntp"

	"go.mongodb.org/mongo-driver/bson"
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

	// Initialize services with NTP time service
	discoveryService := services.NewDiscoveryService(haClient)
	chamberManager := services.NewChamberManager(cfg, db, discoveryService, ntpService)
	registrationService := services.NewRegistrationService(cfg, db, ntpService)
	syncService := services.NewSyncService(cfg, db, ntpService)

	// Use WaitGroups and channels for proper synchronization
	var (
		wg                    sync.WaitGroup
		chamberInitialized    = make(chan struct{})
		registrationCompleted = make(chan struct{})
		executorService       *services.ExecutorService
		executorStarted       = make(chan struct{})
	)

	// Step 1: Wait for Home Assistant connection and initialize chambers
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chamberInitialized)

		for {
			if haClient.IsConnected() {
				log.Printf("Home Assistant connected. Initializing chambers with suffixes: %v", cfg.ChamberSuffixes)

				// Initialize chambers with room separation
				if err := chamberManager.InitializeChambers(ctx); err != nil {
					log.Printf("Warning: Chamber initialization failed: %v", err)
					time.Sleep(10 * time.Second)
					continue
				}

				// Log discovered chambers
				parentChamber := chamberManager.GetParentChamber()
				roomChambers := chamberManager.GetRoomChambers()

				log.Printf("Successfully discovered chambers:")
				log.Printf("  Parent chamber: %s (ID: %s)", parentChamber.Name, parentChamber.ID.Hex())
				for roomSuffix, roomChamber := range roomChambers {
					log.Printf("  Room chamber '%s': %s (%d inputs, %d lamps, %d zones)",
						roomSuffix, roomChamber.Name,
						len(roomChamber.InputNumbers), len(roomChamber.Lamps), len(roomChamber.WateringZones))
				}

				haClient.Status = true

				// Set chamber ID for services immediately after initialization
				registrationService.SetChamberID(parentChamber.ID)
				log.Printf("Services configured with parent chamber: %s", parentChamber.Name)

				// Create executor service now that chamber is ready
				executorService = services.NewExecutorService(db, haClient, parentChamber, ntpService)
				log.Println("Executor service created")

				break
			}
			log.Println("Waiting for Home Assistant connection...")
			time.Sleep(10 * time.Second)
		}
	}()

	// Step 2: Register with backend after chamber initialization
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(registrationCompleted)

		// Wait for chamber initialization
		<-chamberInitialized

		parentChamber := chamberManager.GetParentChamber()
		if parentChamber == nil {
			log.Printf("Parent chamber is nil after initialization")
			return
		}

		log.Println("Registering parent chamber with backend...")
		if err := registrationService.RegisterWithBackend(parentChamber); err != nil {
			log.Printf("Warning: Parent chamber registration failed: %v", err)
			// Don't return here - we can still function without backend registration
		} else {
			// Update sync service with backend ID only if registration was successful
			syncService.SetBackendID(parentChamber.BackendID)
			registrationService.SetBackendID(parentChamber.BackendID)

			// Register room chambers with backend
			log.Println("Registering room chambers with backend...")
			if err := chamberManager.RegisterRoomChambersWithBackend(registrationService); err != nil {
				log.Printf("Warning: Room chambers registration failed: %v", err)
			} else {
				log.Println("✅ All chambers registered successfully with backend")
			}
		}
	}()

	// Step 3: Start background services after registration
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait for registration to complete (or fail)
		<-registrationCompleted

		// Start heartbeat service (it will handle the case where backend ID is not set)
		go func() {
			log.Println("Starting heartbeat service...")
			registrationService.StartHeartbeat(ctx)
		}()

		// Start sync service
		go func() {
			log.Println("Starting sync service...")
			syncService.StartSync(ctx)
		}()
	}()

	// Step 4: Start executor service after all prerequisites are met
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(executorStarted)

		// Wait for chamber initialization
		<-chamberInitialized

		// Wait a bit more to ensure executor service is created
		time.Sleep(2 * time.Second)

		if executorService == nil {
			log.Printf("Executor service is nil - cannot start")
			return
		}

		for {
			if haClient.Status {
				log.Println("Starting executor service...")
				if err := executorService.Start(ctx); err != nil {
					log.Printf("Warning: Failed to start executor service: %v", err)
					time.Sleep(5 * time.Second)
					continue
				} else {
					log.Println("✅ Executor service started successfully")
					break
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	// Start simple HTTP server for health checks
	mux := http.NewServeMux()
	setupRoutes(mux, db, chamberManager, ntpService)

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

	// Stop executor service first
	if executorService != nil {
		log.Println("Stopping executor service...")
		executorService.Stop()
	}

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
func setupRoutes(mux *http.ServeMux, db *database.MongoDB, chamberManager *services.ChamberManager, ntpService *ntp.TimeService) {
	// Health check endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		parentChamber := chamberManager.GetParentChamber()
		roomChambers := chamberManager.GetRoomChambers()

		if parentChamber != nil {
			backendID := "none"
			if !parentChamber.BackendID.IsZero() {
				backendID = parentChamber.BackendID.Hex()
			}
			fmt.Fprintf(w, `{"status":"healthy","chamber_id":"%s","backend_id":"%s","name":"%s","room_chambers":%d,"ntp_enabled":%t,"ntp_connected":%t}`,
				parentChamber.ID.Hex(), backendID, parentChamber.Name, len(roomChambers), ntpService.IsEnabled(), ntpService.IsConnected())
		} else {
			fmt.Fprintf(w, `{"status":"initializing","ntp_enabled":%t,"ntp_connected":%t}`,
				ntpService.IsEnabled(), ntpService.IsConnected())
		}
	})

	// NTP status endpoint
	mux.HandleFunc("/api/v1/ntp/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		status := ntpService.GetStatus()
		statusJSON, _ := json.Marshal(status)
		w.Write(statusJSON)
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

	// Chamber info endpoint
	mux.HandleFunc("/api/v1/chamber", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		parentChamber := chamberManager.GetParentChamber()
		if parentChamber != nil {
			backendID := "none"
			if !parentChamber.BackendID.IsZero() {
				backendID = parentChamber.BackendID.Hex()
			}
			fmt.Fprintf(w, `{"chamber":{"id":"%s","name":"%s","local_ip":"%s","backend_id":"%s","input_numbers":%d,"lamps":%d,"watering_zones":%d}}`,
				parentChamber.ID.Hex(), parentChamber.Name, parentChamber.LocalIP, backendID,
				len(parentChamber.InputNumbers), len(parentChamber.Lamps), len(parentChamber.WateringZones))
		} else {
			fmt.Fprintf(w, `{"error":"Chamber not initialized"}`)
		}
	})

	// Room chambers endpoint
	mux.HandleFunc("/api/v1/chambers/rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		roomChambers := chamberManager.GetRoomChambers()
		fmt.Fprintf(w, `{"room_chambers":[`)

		first := true
		for roomSuffix, roomChamber := range roomChambers {
			if !first {
				fmt.Fprintf(w, ",")
			}
			backendID := "none"
			if !roomChamber.BackendID.IsZero() {
				backendID = roomChamber.BackendID.Hex()
			}
			fmt.Fprintf(w, `{"room_suffix":"%s","id":"%s","name":"%s","backend_id":"%s","input_numbers":%d,"lamps":%d,"watering_zones":%d}`,
				roomSuffix, roomChamber.ID.Hex(), roomChamber.Name, backendID,
				len(roomChamber.InputNumbers), len(roomChamber.Lamps), len(roomChamber.WateringZones))
			first = false
		}

		fmt.Fprintf(w, `]}`)
	})

	// Chambers summary endpoint
	mux.HandleFunc("/api/v1/chambers/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		parentChamber := chamberManager.GetParentChamber()
		roomChambers := chamberManager.GetRoomChambers()

		fmt.Fprintf(w, `{"summary":{"total_chambers":%d,"parent_chamber":`, len(roomChambers)+1)

		if parentChamber != nil {
			fmt.Fprintf(w, `{"name":"%s","entities":%d}`,
				parentChamber.Name,
				len(parentChamber.InputNumbers)+len(parentChamber.Lamps)+len(parentChamber.WateringZones))
		} else {
			fmt.Fprintf(w, `null`)
		}

		fmt.Fprintf(w, `,"room_chambers":[`)
		first := true
		for suffix, chamber := range roomChambers {
			if !first {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, `{"suffix":"%s","name":"%s","entities":%d}`,
				suffix, chamber.Name,
				len(chamber.InputNumbers)+len(chamber.Lamps)+len(chamber.WateringZones))
			first = false
		}
		fmt.Fprintf(w, `]}}`)
	})

	// Experiments endpoint
	mux.HandleFunc("/api/v1/experiments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		parentChamber := chamberManager.GetParentChamber()
		if parentChamber == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"error":"Chamber not initialized"}`)
			return
		}

		ctx := r.Context()
		var filter interface{}
		if !parentChamber.BackendID.IsZero() {
			filter = bson.M{"chamber_id": parentChamber.BackendID}
		} else {
			filter = bson.M{} // Return all experiments if no backend ID
		}

		cursor, err := db.ExperimentsCollection.Find(ctx, filter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"Failed to fetch experiments"}`)
			return
		}
		defer cursor.Close(ctx)

		var experiments []models.Experiment
		if err := cursor.All(ctx, &experiments); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"Failed to decode experiments"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"experiments":%d,"count":%d}`, len(experiments), len(experiments))
	})
}
