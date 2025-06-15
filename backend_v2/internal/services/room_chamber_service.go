package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"backend_v2/internal/config"
	"backend_v2/internal/database"
	"backend_v2/internal/models"
)

// RoomChamberService handles room chamber-related business logic
type RoomChamberService struct {
	db     *database.MongoDB
	config *config.Config
}

// NewRoomChamberService creates a new room chamber service
func NewRoomChamberService(db *database.MongoDB, config *config.Config) *RoomChamberService {
	return &RoomChamberService{
		db:     db,
		config: config,
	}
}

// RegisterRoomChamberRequest represents the request to register a room chamber
type RegisterRoomChamberRequest struct {
	Name            string                `json:"name" binding:"required"`
	RoomSuffix      string                `json:"room_suffix" binding:"required"`
	ParentChamberID string                `json:"parent_chamber_id" binding:"required"`
	Location        string                `json:"location"`
	HAUrl           string                `json:"ha_url" binding:"required"`
	AccessToken     string                `json:"access_token" binding:"required"`
	LocalIP         string                `json:"local_ip" binding:"required"`
	InputNumbers    []models.InputNumber  `json:"input_numbers"`
	Lamps           []models.Lamp         `json:"lamps"`
	WateringZones   []models.WateringZone `json:"watering_zones"`
}

// RegisterRoomChamber registers a new room chamber or updates existing one
func (s *RoomChamberService) RegisterRoomChamber(req *RegisterRoomChamberRequest) (*models.RoomChamber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate parent chamber ID
	parentChamberID, err := primitive.ObjectIDFromHex(req.ParentChamberID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent chamber ID: %v", err)
	}

	// Verify parent chamber exists
	var parentChamber models.Chamber
	err = s.db.ChambersCollection.FindOne(ctx, bson.M{"_id": parentChamberID}).Decode(&parentChamber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("parent chamber not found")
		}
		return nil, fmt.Errorf("failed to find parent chamber: %v", err)
	}

	// Check if room chamber already exists
	var existingRoomChamber models.RoomChamber
	err = s.db.RoomChambersCollection.FindOne(ctx, bson.M{
		"parent_chamber_id": parentChamberID,
		"room_suffix":       req.RoomSuffix,
	}).Decode(&existingRoomChamber)

	now := time.Now()

	if err == mongo.ErrNoDocuments {
		// Create new room chamber
		roomChamber := models.RoomChamber{
			ID:              primitive.NewObjectID(),
			Name:            req.Name,
			RoomSuffix:      req.RoomSuffix,
			ParentChamberID: parentChamberID,
			Location:        req.Location,
			HAUrl:           req.HAUrl,
			AccessToken:     req.AccessToken,
			LocalIP:         req.LocalIP,
			Status:          models.StatusOnline,
			LastHeartbeat:   now,
			InputNumbers:    req.InputNumbers,
			Lamps:           req.Lamps,
			WateringZones:   req.WateringZones,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		_, err = s.db.RoomChambersCollection.InsertOne(ctx, roomChamber)
		if err != nil {
			return nil, fmt.Errorf("failed to create room chamber: %v", err)
		}

		log.Printf("New room chamber registered: %s (%s) for parent %s", roomChamber.Name, roomChamber.RoomSuffix, parentChamber.Name)

		// Log discovered entities
		log.Printf("Room chamber %s entities:", roomChamber.Name)
		log.Printf("  - %d input numbers", len(roomChamber.InputNumbers))
		log.Printf("  - %d lamps", len(roomChamber.Lamps))
		log.Printf("  - %d watering zones", len(roomChamber.WateringZones))

		return &roomChamber, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check existing room chamber: %v", err)
	}

	// Update existing room chamber
	update := bson.M{
		"$set": bson.M{
			"name":           req.Name,
			"location":       req.Location,
			"ha_url":         req.HAUrl,
			"access_token":   req.AccessToken,
			"status":         models.StatusOnline,
			"last_heartbeat": now,
			"input_numbers":  req.InputNumbers,
			"lamps":          req.Lamps,
			"watering_zones": req.WateringZones,
			"updated_at":     now,
		},
	}

	_, err = s.db.RoomChambersCollection.UpdateByID(ctx, existingRoomChamber.ID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update room chamber: %v", err)
	}

	existingRoomChamber.Name = req.Name
	existingRoomChamber.Location = req.Location
	existingRoomChamber.HAUrl = req.HAUrl
	existingRoomChamber.AccessToken = req.AccessToken
	existingRoomChamber.Status = models.StatusOnline
	existingRoomChamber.LastHeartbeat = now
	existingRoomChamber.InputNumbers = req.InputNumbers
	existingRoomChamber.Lamps = req.Lamps
	existingRoomChamber.WateringZones = req.WateringZones
	existingRoomChamber.UpdatedAt = now

	log.Printf("Room chamber updated: %s (%s) for parent %s", existingRoomChamber.Name, existingRoomChamber.RoomSuffix, parentChamber.Name)

	return &existingRoomChamber, nil
}

// UpdateRoomChamberHeartbeat updates the room chamber heartbeat
func (s *RoomChamberService) UpdateRoomChamberHeartbeat(roomChamberID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(roomChamberID)
	if err != nil {
		return fmt.Errorf("invalid room chamber ID: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"last_heartbeat": time.Now(),
			"status":         models.StatusOnline,
		},
	}

	result, err := s.db.RoomChambersCollection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("room chamber not found")
	}

	return nil
}

// GetRoomChamber retrieves a room chamber by ID
func (s *RoomChamberService) GetRoomChamber(roomChamberID string) (*models.RoomChamberResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(roomChamberID)
	if err != nil {
		return nil, fmt.Errorf("invalid room chamber ID: %v", err)
	}

	var roomChamber models.RoomChamber
	err = s.db.RoomChambersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&roomChamber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("room chamber not found")
		}
		return nil, fmt.Errorf("failed to get room chamber: %v", err)
	}

	// Get parent chamber name
	var parentChamber models.Chamber
	parentChamberName := ""
	if err := s.db.ChambersCollection.FindOne(ctx, bson.M{"_id": roomChamber.ParentChamberID}).Decode(&parentChamber); err == nil {
		parentChamberName = parentChamber.Name
	}

	response := &models.RoomChamberResponse{
		ID:                roomChamber.ID,
		Name:              roomChamber.Name,
		RoomSuffix:        roomChamber.RoomSuffix,
		ParentChamberID:   roomChamber.ParentChamberID,
		ParentChamberName: parentChamberName,
		Location:          roomChamber.Location,
		HAUrl:             roomChamber.HAUrl,
		LocalIP:           roomChamber.LocalIP,
		Status:            roomChamber.Status,
		LastHeartbeat:     roomChamber.LastHeartbeat,
		InputNumbers:      roomChamber.InputNumbers,
		Lamps:             roomChamber.Lamps,
		WateringZones:     roomChamber.WateringZones,
		CreatedAt:         roomChamber.CreatedAt,
		UpdatedAt:         roomChamber.UpdatedAt,
	}

	return response, nil
}

// GetRoomChambers retrieves all room chambers
func (s *RoomChamberService) GetRoomChambers() ([]models.RoomChamberResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation pipeline to join with chambers collection
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "chambers",
				"localField":   "parent_chamber_id",
				"foreignField": "_id",
				"as":           "parent_chamber",
			},
		},
		{
			"$addFields": bson.M{
				"parent_chamber_name": bson.M{
					"$arrayElemAt": []interface{}{"$parent_chamber.name", 0},
				},
			},
		},
		{
			"$project": bson.M{
				"parent_chamber": 0, // Remove the joined array
			},
		},
		{
			"$sort": bson.M{"created_at": -1},
		},
	}

	cursor, err := s.db.RoomChambersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get room chambers: %v", err)
	}
	defer cursor.Close(ctx)

	var roomChambers []models.RoomChamberResponse
	if err = cursor.All(ctx, &roomChambers); err != nil {
		return nil, fmt.Errorf("failed to decode room chambers: %v", err)
	}

	return roomChambers, nil
}

// GetRoomChambersByParent retrieves all room chambers for a specific parent chamber
func (s *RoomChamberService) GetRoomChambersByParent(parentChamberID string) ([]models.RoomChamberResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(parentChamberID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent chamber ID: %v", err)
	}

	// Aggregation pipeline to join with chambers collection
	pipeline := []bson.M{
		{
			"$match": bson.M{"parent_chamber_id": objectID},
		},
		{
			"$lookup": bson.M{
				"from":         "chambers",
				"localField":   "parent_chamber_id",
				"foreignField": "_id",
				"as":           "parent_chamber",
			},
		},
		{
			"$addFields": bson.M{
				"parent_chamber_name": bson.M{
					"$arrayElemAt": []interface{}{"$parent_chamber.name", 0},
				},
			},
		},
		{
			"$project": bson.M{
				"parent_chamber": 0, // Remove the joined array
			},
		},
		{
			"$sort": bson.M{"room_suffix": 1},
		},
	}

	cursor, err := s.db.RoomChambersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get room chambers: %v", err)
	}
	defer cursor.Close(ctx)

	var roomChambers []models.RoomChamberResponse
	if err = cursor.All(ctx, &roomChambers); err != nil {
		return nil, fmt.Errorf("failed to decode room chambers: %v", err)
	}

	return roomChambers, nil
}

// UpdateRoomChamberStatus updates the status of room chambers based on heartbeat timeout
func (s *RoomChamberService) UpdateRoomChamberStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find room chambers that haven't sent heartbeat within timeout period
	cutoffTime := time.Now().Add(-s.config.HeartbeatTimeout)

	filter := bson.M{
		"status":         models.StatusOnline,
		"last_heartbeat": bson.M{"$lt": cutoffTime},
	}

	update := bson.M{
		"$set": bson.M{
			"status": models.StatusOffline,
		},
	}

	result, err := s.db.RoomChambersCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update room chamber status: %v", err)
	}

	if result.ModifiedCount > 0 {
		log.Printf("Marked %d room chambers as offline", result.ModifiedCount)
	}

	return nil
}

// GetRoomChamberWateringZones handles GET /room-chambers/:id/watering-zones
func (s *RoomChamberService) GetRoomChamberWateringZones(roomChamberID string) (interface{}, error) {
	roomChamber, err := s.GetRoomChamber(roomChamberID)
	if err != nil {
		return nil, err
	}

	// Build response with watering zones and their associated input numbers
	type WateringZoneResponse struct {
		Zone         models.WateringZone            `json:"zone"`
		InputNumbers map[string]*models.InputNumber `json:"input_numbers"`
	}

	var response []WateringZoneResponse

	// Create a temporary room chamber struct to use the method
	tempRoomChamber := &models.RoomChamber{
		InputNumbers:  roomChamber.InputNumbers,
		WateringZones: roomChamber.WateringZones,
	}

	wateringInputs := tempRoomChamber.GetWateringInputNumbers()

	for _, zone := range roomChamber.WateringZones {
		if inputs, ok := wateringInputs[zone.Name]; ok {
			response = append(response, WateringZoneResponse{
				Zone:         zone,
				InputNumbers: inputs,
			})
		}
	}

	return response, nil
}
