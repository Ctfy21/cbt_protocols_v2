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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

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

	// Get or create chamber
	chamber, err := getOrCreateChamber(ctx, db, cfg)
	if err != nil {
		log.Fatalf("Failed to get or create chamber: %v", err)
	}

	log.Printf("Chamber initialized: %s (ID: %s)", chamber.Name, chamber.ID.Hex())

	// Initialize services
	discoveryService := services.NewDiscoveryService(haClient)
	registrationService := services.NewRegistrationService(cfg, db)
	syncService := services.NewSyncService(cfg, db)
	executorService := services.NewExecutorService(db, haClient, chamber)

	// Run initial discovery
	log.Println("Running initial entity discovery...")
	go func() {
		for {
			if haClient.IsConnected() {
				inputNumbers, lamps, wateringZones, err := discoveryService.DiscoverInputNumbers()

				for _, inputNumber := range inputNumbers {
					log.Printf("Input number: %s, entity_id: %s", inputNumber.Name, inputNumber.EntityID)
				}

				if err != nil {
					log.Printf("Warning: Entity discovery failed: %v", err)
				} else {
					// Update chamber with discovered entities
					chamber.InputNumbers = inputNumbers
					chamber.Lamps = lamps
					chamber.WateringZones = wateringZones

					// Save discovered entities
					if err := saveDiscoveredEntities(ctx, db, chamber); err != nil {
						log.Printf("Warning: Failed to save discovered entities: %v", err)
					}

					log.Printf("Discovered: %d input numbers, %d lamps, %d watering zones",
						len(inputNumbers), len(lamps), len(wateringZones))
					haClient.Status = true
					break
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	// Set chamber ID for services
	registrationService.SetChamberID(chamber.ID)

	// Register with backend
	log.Println("Registering with backend...")
	if err := registrationService.RegisterWithBackend(chamber); err != nil {
		log.Printf("Warning: Initial registration failed: %v", err)
	} else {
		// Update sync service with backend ID
		syncService.SetBackendID(chamber.BackendID)
		registrationService.SetBackendID(chamber.BackendID)
	}

	// Start background services
	go registrationService.StartHeartbeat(ctx)
	go syncService.StartSync(ctx)

	// Start executor service
	go func() {
		for {
			if haClient.Status {
				if err := executorService.Start(ctx); err != nil {
					log.Printf("Warning: Failed to start executor service: %v", err)
				}
				break
			}
			time.Sleep(10 * time.Second)
		}
	}()
	defer executorService.Stop()

	// Start simple HTTP server for health checks
	mux := http.NewServeMux()
	setupRoutes(mux, db, chamber)

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
func setupRoutes(mux *http.ServeMux, db *database.MongoDB, chamber *models.Chamber) {
	// Health check endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","chamber_id":"%s","backend_id":"%s","name":"%s"}`,
			chamber.ID.Hex(), chamber.BackendID.Hex(), chamber.Name)
	})

	// Chamber info endpoint
	mux.HandleFunc("/api/v1/chamber", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"chamber":{"id":"%s","name":"%s","local_ip":"%s","backend_id":"%s","input_numbers":%d,"lamps":%d,"watering_zones":%d}}`,
			chamber.ID.Hex(), chamber.Name, chamber.LocalIP, chamber.BackendID.Hex(),
			len(chamber.InputNumbers), len(chamber.Lamps), len(chamber.WateringZones))
	})

	// Experiments endpoint
	mux.HandleFunc("/api/v1/experiments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		cursor, err := db.ExperimentsCollection.Find(ctx, bson.M{"chamber_id": chamber.BackendID})
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

// getOrCreateChamber retrieves existing chamber or creates a new one
func getOrCreateChamber(ctx context.Context, db *database.MongoDB, cfg *config.Config) (*models.Chamber, error) {
	// Try to find existing chamber
	var chamber models.Chamber
	err := db.ChambersCollection.FindOne(ctx, bson.M{}).Decode(&chamber)

	if err == nil {
		// Chamber exists, return it
		log.Printf("Found existing chamber: %s", chamber.Name)
		return &chamber, nil
	}

	// Create new chamber
	log.Println("Creating new chamber...")
	chamber = models.Chamber{
		ID:               primitive.NewObjectID(),
		Name:             cfg.ChamberName,
		LocalIP:          cfg.LocalIP,
		HomeAssistantURL: cfg.HomeAssistantURL,
		InputNumbers:     []models.InputNumber{},
		Lamps:            []models.Lamp{},
		WateringZones:    []models.WateringZone{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Insert chamber
	_, err = db.ChambersCollection.InsertOne(ctx, chamber)
	if err != nil {
		return nil, fmt.Errorf("failed to create chamber: %w", err)
	}

	log.Printf("Created new chamber: %s", chamber.Name)
	return &chamber, nil
}

// saveDiscoveredEntities saves the discovered entities to the database
func saveDiscoveredEntities(ctx context.Context, db *database.MongoDB, chamber *models.Chamber) error {
	update := bson.M{
		"$set": bson.M{
			"input_numbers":  chamber.InputNumbers,
			"lamps":          chamber.Lamps,
			"watering_zones": chamber.WateringZones,
			"updated_at":     time.Now(),
		},
	}

	_, err := db.ChambersCollection.UpdateOne(
		ctx,
		bson.M{"_id": chamber.ID},
		update,
	)

	return err
}
