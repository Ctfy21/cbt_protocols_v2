package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local_api_v2/internal/config"
	"local_api_v2/internal/database"
	"local_api_v2/internal/models"
	"local_api_v2/internal/services"
	"local_api_v2/pkg/homeassistant"

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
	haClient.Status = haClient.IsConnected()

	// Initialize services
	discoveryService := services.NewDiscoveryService(haClient)

	// Chamber manager for handling multiple chambers with custom suffixes
	chamberManager := services.NewChamberManager(cfg, db, discoveryService)

	registrationService := services.NewRegistrationService(cfg, db)
	syncService := services.NewSyncService(cfg, db)

	// Run initial discovery and chamber initialization
	log.Println("Running initial entity discovery and chamber initialization...")
	go func() {
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
				break
			}
			log.Println("Waiting for Home Assistant connection...")
			time.Sleep(10 * time.Second)
		}
	}()

	// Wait for chamber initialization and set chamber ID for services
	go func() {
		for {
			parentChamber := chamberManager.GetParentChamber()
			if parentChamber != nil {
				registrationService.SetChamberID(parentChamber.ID)
				log.Printf("Services configured with parent chamber: %s", parentChamber.Name)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Create executor service after chamber is initialized
	var executorService *services.ExecutorService
	go func() {
		for {
			parentChamber := chamberManager.GetParentChamber()
			if parentChamber != nil {
				executorService = services.NewExecutorService(db, haClient, parentChamber)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Register with backend
	go func() {
		for {
			parentChamber := chamberManager.GetParentChamber()
			if parentChamber != nil {
				log.Println("Registering parent chamber with backend...")
				if err := registrationService.RegisterWithBackend(parentChamber); err != nil {
					log.Printf("Warning: Parent chamber registration failed: %v", err)
				} else {
					// Update sync service with backend ID
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
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Start background services
	go registrationService.StartHeartbeat(ctx)
	go syncService.StartSync(ctx)

	// Start executor service
	go func() {
		for {
			if haClient.Status {
				if err := executorService.Start(ctx); err != nil {
					log.Printf("Warning: Failed to start executor service: %v", err)
				} else {
					log.Println("✅ Executor service started successfully")
				}
				break
			}
			time.Sleep(10 * time.Second)
		}
	}()
	defer executorService.Stop()

	// Start simple HTTP server for health checks
	mux := http.NewServeMux()
	setupRoutes(mux, db, chamberManager)

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
	<-quit

	log.Println("Shutting down server...")

	// Shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupRoutes configures HTTP routes
func setupRoutes(mux *http.ServeMux, db *database.MongoDB, chamberManager *services.ChamberManager) {
	// Health check endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		parentChamber := chamberManager.GetParentChamber()
		roomChambers := chamberManager.GetRoomChambers()

		if parentChamber != nil {
			fmt.Fprintf(w, `{"status":"healthy","chamber_id":"%s","backend_id":"%s","name":"%s","room_chambers":%d}`,
				parentChamber.ID.Hex(), parentChamber.BackendID.Hex(), parentChamber.Name, len(roomChambers))
		} else {
			fmt.Fprintf(w, `{"status":"initializing"}`)
		}
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
			fmt.Fprintf(w, `{"chamber":{"id":"%s","name":"%s","local_ip":"%s","backend_id":"%s","input_numbers":%d,"lamps":%d,"watering_zones":%d}}`,
				parentChamber.ID.Hex(), parentChamber.Name, parentChamber.LocalIP, parentChamber.BackendID.Hex(),
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
			backendID := ""
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
		cursor, err := db.ExperimentsCollection.Find(ctx, bson.M{"chamber_id": parentChamber.BackendID})
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
