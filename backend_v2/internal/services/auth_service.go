package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"backend_v2/internal/config"
	"backend_v2/internal/database"
	"backend_v2/internal/models"
)

// AuthService handles authentication logic
type AuthService struct {
	db     *database.MongoDB
	config *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(db *database.MongoDB, config *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: config,
	}
}

// Register creates a new user
// func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	// Check if user already exists
// 	var existingUser models.User
// 	err := s.db.UsersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
// 	if err == nil {
// 		return nil, fmt.Errorf("user with email %s already exists", req.Email)
// 	}
// 	if err != mongo.ErrNoDocuments {
// 		return nil, fmt.Errorf("failed to check existing user: %v", err)
// 	}

// 	// Hash password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to hash password: %v", err)
// 	}

// 	// Create user
// 	now := time.Now()
// 	user := models.User{
// 		ID:        primitive.NewObjectID(),
// 		Email:     req.Email,
// 		Password:  string(hashedPassword),
// 		Name:      req.Name,
// 		Role:      models.RoleUser,
// 		IsActive:  true,
// 		CreatedAt: now,
// 		UpdatedAt: now,
// 	}

// 	_, err = s.db.UsersCollection.InsertOne(ctx, user)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create user: %v", err)
// 	}

// 	// Generate tokens
// 	authResp, err := s.generateAuthResponse(&user, ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create session record for new user
// 	session := models.Session{
// 		ID:           primitive.NewObjectID(),
// 		UserID:       user.ID,
// 		Token:        authResp.Token,
// 		RefreshToken: authResp.RefreshToken,
// 		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
// 		CreatedAt:    now,
// 		UserAgent:    "", // No user agent available during registration
// 		IP:           "", // No IP available during registration
// 	}

// 	_, err = s.db.SessionsCollection.InsertOne(ctx, session)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create session: %v", err)
// 	}

// 	return authResp, nil
// }

// Login authenticates a user
func (s *AuthService) Login(req *models.LoginRequest, userAgent, ip string) (*models.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email
	var user models.User
	err := s.db.UsersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update last login
	now := time.Now()
	_, err = s.db.UsersCollection.UpdateByID(ctx, user.ID, bson.M{
		"$set": bson.M{
			"last_login": now,
			"updated_at": now,
		},
	})
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Generate tokens and create session
	authResp, err := s.generateAuthResponse(&user, ctx)
	if err != nil {
		return nil, err
	}

	// Create session record
	session := models.Session{
		ID:           primitive.NewObjectID(),
		UserID:       user.ID,
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
		CreatedAt:    now,
		UserAgent:    userAgent,
		IP:           ip,
	}

	_, err = s.db.SessionsCollection.InsertOne(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return authResp, nil
}

// RefreshToken refreshes the authentication token
func (s *AuthService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find session by refresh token
	var session models.Session
	err := s.db.SessionsCollection.FindOne(ctx, bson.M{"refresh_token": refreshToken}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("invalid refresh token")
		}
		return nil, fmt.Errorf("failed to find session: %v", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt.Add(30 * 24 * time.Hour)) { // Refresh tokens valid for 30 days
		return nil, fmt.Errorf("refresh token expired")
	}

	// Find user
	var user models.User
	err = s.db.UsersCollection.FindOne(ctx, bson.M{"_id": session.UserID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Delete old session
	_, err = s.db.SessionsCollection.DeleteOne(ctx, bson.M{"_id": session.ID})
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to delete old session: %v\n", err)
	}

	// Generate new tokens
	authResp, err := s.generateAuthResponse(&user, ctx)
	if err != nil {
		return nil, err
	}

	// Create new session record
	newSession := models.Session{
		ID:           primitive.NewObjectID(),
		UserID:       user.ID,
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
		CreatedAt:    time.Now(),
		UserAgent:    session.UserAgent, // Preserve original user agent
		IP:           session.IP,        // Preserve original IP
	}

	_, err = s.db.SessionsCollection.InsertOne(ctx, newSession)
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %v", err)
	}

	return authResp, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete session
	_, err := s.db.SessionsCollection.DeleteOne(ctx, bson.M{"token": token})
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in token")
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	// Check if session exists
	var session models.Session
	err = s.db.SessionsCollection.FindOne(ctx, bson.M{"token": tokenString}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session: %v", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Find user
	var user models.User
	err = s.db.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var user models.User
	err = s.db.UsersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (s *AuthService) UpdateUser(userID string, update bson.M) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	// Add updated_at timestamp
	update["updated_at"] = time.Now()

	// Update user
	result := s.db.UsersCollection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var user models.User
	err = result.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	return &user, nil
}

// Helper functions

func (s *AuthService) generateAuthResponse(user *models.User, ctx context.Context) (*models.AuthResponse, error) {
	// Generate JWT token
	expiresIn := int64(s.config.JWTExpiration.Seconds())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(s.config.JWTExpiration).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	// Clear password before returning
	user.Password = ""

	return &models.AuthResponse{
		User:         *user,
		Token:        tokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VerifyPassword verifies a user's password
func (s *AuthService) VerifyPassword(userID string, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	var user models.User
	err = s.db.UsersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to find user: %v", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}

// UpdatePassword updates a user's password
func (s *AuthService) UpdatePassword(userID string, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password
	_, err = s.db.UsersCollection.UpdateByID(ctx, objectID, bson.M{
		"$set": bson.M{
			"password":   string(hashedPassword),
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	// Invalidate all sessions for this user
	_, err = s.db.SessionsCollection.DeleteMany(ctx, bson.M{"user_id": objectID})
	if err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to invalidate sessions: %v\n", err)
	}

	return nil
}

// CreateUser creates a new user (admin function)
func (s *AuthService) CreateUser(req *models.RegisterRequest, role models.UserRole) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingUser models.User
	err := s.db.UsersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}
	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		Role:      role,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = s.db.UsersCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Clear password before returning
	user.Password = ""

	return &user, nil
}

// GetAllUsers retrieves all users (admin function)
func (s *AuthService) GetAllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.UsersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %v", err)
	}

	// Clear passwords from all users
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}
