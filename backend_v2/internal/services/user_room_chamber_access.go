package services

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"backend_v2/internal/database"
	"backend_v2/internal/models"
)

// UserRoomChamberAccessService handles user room chamber access operations
type UserRoomChamberAccessService struct {
	db *database.MongoDB
}

// NewUserRoomChamberAccessService creates a new user room chamber access service
func NewUserRoomChamberAccessService(db *database.MongoDB) *UserRoomChamberAccessService {
	return &UserRoomChamberAccessService{
		db: db,
	}
}

// GrantRoomChamberAccess grants room chamber access to a user
func (s *UserRoomChamberAccessService) GrantRoomChamberAccess(userID, roomChamberID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if access already exists
	count, err := s.db.UserRoomChamberAccessCollection.CountDocuments(ctx, bson.M{
		"user_id":         userID,
		"room_chamber_id": roomChamberID,
	})
	if err != nil {
		return fmt.Errorf("failed to check existing access: %v", err)
	}

	if count > 0 {
		return nil // Access already exists
	}

	access := models.UserRoomChamberAccess{
		ID:            primitive.NewObjectID(),
		UserID:        userID,
		RoomChamberID: roomChamberID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = s.db.UserRoomChamberAccessCollection.InsertOne(ctx, access)
	if err != nil {
		return fmt.Errorf("failed to grant room chamber access: %v", err)
	}

	return nil
}

// RevokeRoomChamberAccess revokes room chamber access from a user
func (s *UserRoomChamberAccessService) RevokeRoomChamberAccess(userID, roomChamberID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.db.UserRoomChamberAccessCollection.DeleteMany(ctx, bson.M{
		"user_id":         userID,
		"room_chamber_id": roomChamberID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke room chamber access: %v", err)
	}

	return nil
}

// SetUserRoomChamberAccess sets room chamber access for a user (replaces all existing access)
func (s *UserRoomChamberAccessService) SetUserRoomChamberAccess(userIDStr string, roomChamberIDStrs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Convert room chamber ID strings to ObjectIDs
	roomChamberIDs := make([]primitive.ObjectID, len(roomChamberIDStrs))
	for i, idStr := range roomChamberIDStrs {
		roomChamberID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return fmt.Errorf("invalid room chamber ID %s: %v", idStr, err)
		}
		roomChamberIDs[i] = roomChamberID
	}

	// Start transaction
	session, err := s.db.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Remove all existing access for this user
		_, err := s.db.UserRoomChamberAccessCollection.DeleteMany(sessCtx, bson.M{
			"user_id": userID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing access: %v", err)
		}

		// Add new access records
		if len(roomChamberIDs) > 0 {
			var accessRecords []interface{}
			now := time.Now()
			for _, roomChamberID := range roomChamberIDs {
				accessRecords = append(accessRecords, models.UserRoomChamberAccess{
					ID:            primitive.NewObjectID(),
					UserID:        userID,
					RoomChamberID: roomChamberID,
					CreatedAt:     now,
					UpdatedAt:     now,
				})
			}

			_, err = s.db.UserRoomChamberAccessCollection.InsertMany(sessCtx, accessRecords)
			if err != nil {
				return nil, fmt.Errorf("failed to insert new access records: %v", err)
			}
		}

		return nil, nil
	})

	return err
}

// GetUserRoomChamberAccess gets all room chambers a user has access to
func (s *UserRoomChamberAccessService) GetUserRoomChamberAccess(userID primitive.ObjectID) ([]models.RoomChamberResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get room chamber IDs user has access to
	cursor, err := s.db.UserRoomChamberAccessCollection.Find(ctx, bson.M{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user access: %v", err)
	}
	defer cursor.Close(ctx)

	var accessRecords []models.UserRoomChamberAccess
	if err := cursor.All(ctx, &accessRecords); err != nil {
		return nil, fmt.Errorf("failed to decode access records: %v", err)
	}

	if len(accessRecords) == 0 {
		return []models.RoomChamberResponse{}, nil
	}

	// Get room chamber IDs
	roomChamberIDs := make([]primitive.ObjectID, len(accessRecords))
	for i, access := range accessRecords {
		roomChamberIDs[i] = access.RoomChamberID
	}

	// Aggregation pipeline to get room chambers with parent chamber names
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id": bson.M{"$in": roomChamberIDs},
			},
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
	}

	cursor, err = s.db.RoomChambersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get room chambers: %v", err)
	}
	defer cursor.Close(ctx)

	var roomChambers []models.RoomChamberResponse
	if err := cursor.All(ctx, &roomChambers); err != nil {
		return nil, fmt.Errorf("failed to decode room chambers: %v", err)
	}

	return roomChambers, nil
}

// HasRoomChamberAccess checks if user has access to specific room chamber
func (s *UserRoomChamberAccessService) HasRoomChamberAccess(userID, roomChamberID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.db.UserRoomChamberAccessCollection.CountDocuments(ctx, bson.M{
		"user_id":         userID,
		"room_chamber_id": roomChamberID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check room chamber access: %v", err)
	}

	return count > 0, nil
}

// GetAllUsersWithRoomChamberAccess gets all users with their room chamber access
func (s *UserRoomChamberAccessService) GetAllUsersWithRoomChamberAccess() ([]models.UserWithRoomChamberAccess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all users
	cursor, err := s.db.UsersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %v", err)
	}

	// Get room chamber access for each user
	var result []models.UserWithRoomChamberAccess
	for _, user := range users {
		roomChambers, err := s.GetUserRoomChamberAccess(user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get room chambers for user %s: %v", user.ID.Hex(), err)
		}

		result = append(result, models.UserWithRoomChamberAccess{
			User:         user,
			RoomChambers: roomChambers,
		})
	}

	return result, nil
}
