package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// User represents a user in the system
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"` // Never expose password in JSON
	Name      string             `bson:"name" json:"name"`
	Role      UserRole           `bson:"role" json:"role"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	LastLogin *time.Time         `bson:"last_login,omitempty" json:"last_login,omitempty"`
}

// Session represents a user session
type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token        string             `bson:"token" json:"token"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	IP           string             `bson:"ip" json:"ip"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
