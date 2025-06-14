package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// APITokenType represents the type of API token
type APITokenType string

const (
	APITokenTypeService  APITokenType = "service"
	APITokenTypePersonal APITokenType = "personal"
)

// APIToken represents an API token for service-to-service authentication
type APIToken struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Token       string             `bson:"token" json:"token"` // The actual token
	Type        APITokenType       `bson:"type" json:"type"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`           // User who created it
	ServiceName string             `bson:"service_name" json:"service_name"` // For service tokens
	Permissions []string           `bson:"permissions" json:"permissions"`   // List of permissions
	IsActive    bool               `bson:"is_active" json:"is_active"`
	ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	LastUsedAt  *time.Time         `bson:"last_used_at,omitempty" json:"last_used_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateAPITokenRequest represents the request to create an API token
type CreateAPITokenRequest struct {
	Name        string       `json:"name" binding:"required"`
	Type        APITokenType `json:"type" binding:"required"`
	ServiceName string       `json:"service_name,omitempty"`
	Permissions []string     `json:"permissions"`
	ExpiresAt   string       `json:"expires_at,omitempty"`
}

// APITokenResponse represents the response when creating an API token
type APITokenResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Token       string             `json:"token"`
	Type        APITokenType       `json:"type"`
	ServiceName string             `json:"service_name,omitempty"`
	Permissions []string           `json:"permissions"`
	IsActive    bool               `json:"is_active"`
	ExpiresAt   string             `json:"expires_at,omitempty"`
	CreatedAt   string             `json:"created_at"`
}
