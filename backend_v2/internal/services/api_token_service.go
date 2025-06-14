package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"backend_v2/internal/database"
	"backend_v2/internal/models"
)

// APITokenService handles API token operations
type APITokenService struct {
	db *database.MongoDB
}

// NewAPITokenService creates a new API token service
func NewAPITokenService(db *database.MongoDB) *APITokenService {
	return &APITokenService{
		db: db,
	}
}

func (s *APITokenService) ParseExpiresAt(expiresAt string) (*time.Time, error) {
	i, err := strconv.ParseInt(expiresAt, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at format: %v", err)
	}
	t := time.Unix(i, 0)
	return &t, nil
}

// CreateAPIToken creates a new API token
func (s *APITokenService) CreateAPIToken(userID primitive.ObjectID, req *models.CreateAPITokenRequest) (*models.APITokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse expiration if provided
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		parsed, err := s.ParseExpiresAt(req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %v", err)
		}
		expiresAt = parsed
	}

	// Generate secure random token
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Create API token
	now := time.Now()
	apiToken := models.APIToken{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Token:       token,
		Type:        req.Type,
		UserID:      userID,
		ServiceName: req.ServiceName,
		Permissions: req.Permissions,
		IsActive:    true,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Insert into database
	_, err = s.db.APITokensCollection.InsertOne(ctx, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create API token: %v", err)
	}

	// Prepare response
	response := &models.APITokenResponse{
		ID:          apiToken.ID,
		Name:        apiToken.Name,
		Token:       apiToken.Token,
		Type:        apiToken.Type,
		ServiceName: apiToken.ServiceName,
		Permissions: apiToken.Permissions,
		IsActive:    apiToken.IsActive,
		CreatedAt:   apiToken.CreatedAt.Format(time.RFC3339),
	}

	if apiToken.ExpiresAt != nil {
		response.ExpiresAt = apiToken.ExpiresAt.Format(time.RFC3339)
	}

	return response, nil
}

// ValidateAPIToken validates an API token and returns associated user
func (s *APITokenService) ValidateAPIToken(token string) (*models.User, *models.APIToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find API token
	var apiToken models.APIToken
	err := s.db.APITokensCollection.FindOne(ctx, bson.M{
		"token":     token,
		"is_active": true,
	}).Decode(&apiToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil, fmt.Errorf("invalid API token")
		}
		return nil, nil, fmt.Errorf("failed to find API token: %v", err)
	}

	// Check if token is expired
	if apiToken.ExpiresAt != nil && time.Now().After(*apiToken.ExpiresAt) {
		return nil, nil, fmt.Errorf("API token expired")
	}

	// Update last used timestamp
	now := time.Now()
	_, err = s.db.APITokensCollection.UpdateOne(ctx, bson.M{"_id": apiToken.ID}, bson.M{
		"$set": bson.M{
			"last_used_at": now,
			"updated_at":   now,
		},
	})
	if err != nil {
		// Log but don't fail validation
		fmt.Printf("Failed to update API token last used: %v\n", err)
	}

	// Get associated user
	var user models.User
	err = s.db.UsersCollection.FindOne(ctx, bson.M{"_id": apiToken.UserID}).Decode(&user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find associated user: %v", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, fmt.Errorf("associated user account is deactivated")
	}

	return &user, &apiToken, nil
}

// GetAPITokens gets all API tokens for a user
func (s *APITokenService) GetAPITokens(userID primitive.ObjectID) ([]models.APIToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.APITokensCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find API tokens: %v", err)
	}
	defer cursor.Close(ctx)

	var tokens []models.APIToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, fmt.Errorf("failed to decode API tokens: %v", err)
	}

	// Remove sensitive token data from response
	// for i := range tokens {
	// 	tokens[i].Token = "" // Don't expose actual token
	// }

	return tokens, nil
}

// RevokeAPIToken revokes an API token
func (s *APITokenService) RevokeAPIToken(tokenID primitive.ObjectID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := s.db.APITokensCollection.UpdateOne(ctx, bson.M{
		"_id":     tokenID,
		"user_id": userID,
	}, bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to revoke API token: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("API token not found or not owned by user")
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func (s *APITokenService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "svc_" + hex.EncodeToString(bytes), nil
}
