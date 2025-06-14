package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserChamberAccess represents user access to chambers
type UserChamberAccess struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ChamberID primitive.ObjectID `bson:"chamber_id" json:"chamber_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserChamberAccessRequest represents request to grant/revoke chamber access
type UserChamberAccessRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
	ChamberIDs []string `json:"chamber_ids" binding:"required"`
}

// UserWithChamberAccess represents user with their chamber access
type UserWithChamberAccess struct {
	User     User      `json:"user"`
	Chambers []Chamber `json:"chambers"`
}
