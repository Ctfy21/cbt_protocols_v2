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

// UserChamberAccessService handles user chamber access operations
type UserChamberAccessService struct {
	db *database.MongoDB
}

// NewUserChamberAccessService creates a new user chamber access service
func NewUserChamberAccessService(db *database.MongoDB) *UserChamberAccessService {
	return &UserChamberAccessService{
		db: db,
	}
}

// GrantChamberAccess grants chamber access to a user
func (s *UserChamberAccessService) GrantChamberAccess(userID, chamberID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if access already exists
	count, err := s.db.UserChamberAccessCollection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"chamber_id": chamberID,
	})
	if err != nil {
		return fmt.Errorf("failed to check existing access: %v", err)
	}

	if count > 0 {
		return nil // Access already exists
	}

	access := models.UserChamberAccess{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ChamberID: chamberID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.db.UserChamberAccessCollection.InsertOne(ctx, access)
	if err != nil {
		return fmt.Errorf("failed to grant chamber access: %v", err)
	}

	return nil
}

// RevokeChamberAccess revokes chamber access from a user
func (s *UserChamberAccessService) RevokeChamberAccess(userID, chamberID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.db.UserChamberAccessCollection.DeleteMany(ctx, bson.M{
		"user_id":    userID,
		"chamber_id": chamberID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke chamber access: %v", err)
	}

	return nil
}

// SetUserChamberAccess sets chamber access for a user (replaces all existing access)
func (s *UserChamberAccessService) SetUserChamberAccess(userIDStr string, chamberIDStrs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Convert chamber ID strings to ObjectIDs
	chamberIDs := make([]primitive.ObjectID, len(chamberIDStrs))
	for i, idStr := range chamberIDStrs {
		chamberID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return fmt.Errorf("invalid chamber ID %s: %v", idStr, err)
		}
		chamberIDs[i] = chamberID
	}

	// Start transaction
	session, err := s.db.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Remove all existing access for this user
		_, err := s.db.UserChamberAccessCollection.DeleteMany(sessCtx, bson.M{
			"user_id": userID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing access: %v", err)
		}

		// Add new access records
		if len(chamberIDs) > 0 {
			var accessRecords []interface{}
			now := time.Now()
			for _, chamberID := range chamberIDs {
				accessRecords = append(accessRecords, models.UserChamberAccess{
					ID:        primitive.NewObjectID(),
					UserID:    userID,
					ChamberID: chamberID,
					CreatedAt: now,
					UpdatedAt: now,
				})
			}

			_, err = s.db.UserChamberAccessCollection.InsertMany(sessCtx, accessRecords)
			if err != nil {
				return nil, fmt.Errorf("failed to insert new access records: %v", err)
			}
		}

		return nil, nil
	})

	return err
}

// GetUserChamberAccess gets all chambers a user has access to
func (s *UserChamberAccessService) GetUserChamberAccess(userID primitive.ObjectID) ([]models.Chamber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get chamber IDs user has access to
	cursor, err := s.db.UserChamberAccessCollection.Find(ctx, bson.M{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user access: %v", err)
	}
	defer cursor.Close(ctx)

	var accessRecords []models.UserChamberAccess
	if err := cursor.All(ctx, &accessRecords); err != nil {
		return nil, fmt.Errorf("failed to decode access records: %v", err)
	}

	if len(accessRecords) == 0 {
		return []models.Chamber{}, nil
	}

	// Get chamber IDs
	chamberIDs := make([]primitive.ObjectID, len(accessRecords))
	for i, access := range accessRecords {
		chamberIDs[i] = access.ChamberID
	}

	// Get chambers
	cursor, err = s.db.ChambersCollection.Find(ctx, bson.M{
		"_id": bson.M{"$in": chamberIDs},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get chambers: %v", err)
	}
	defer cursor.Close(ctx)

	var chambers []models.Chamber
	if err := cursor.All(ctx, &chambers); err != nil {
		return nil, fmt.Errorf("failed to decode chambers: %v", err)
	}

	return chambers, nil
}

// GetAllUsersWithChamberAccess gets all users with their chamber access
func (s *UserChamberAccessService) GetAllUsersWithChamberAccess() ([]models.UserWithChamberAccess, error) {
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

	// Get chamber access for each user
	var result []models.UserWithChamberAccess
	for _, user := range users {
		chambers, err := s.GetUserChamberAccess(user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chambers for user %s: %v", user.ID.Hex(), err)
		}

		result = append(result, models.UserWithChamberAccess{
			User:     user,
			Chambers: chambers,
		})
	}

	return result, nil
}

// HasChamberAccess checks if user has access to specific chamber
func (s *UserChamberAccessService) HasChamberAccess(userID, chamberID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.db.UserChamberAccessCollection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"chamber_id": chamberID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check chamber access: %v", err)
	}

	return count > 0, nil
}
