package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRoomChamberAccess represents user access to room chambers
type UserRoomChamberAccess struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	RoomChamberID primitive.ObjectID `bson:"room_chamber_id" json:"room_chamber_id"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserWithRoomChamberAccess represents user with their room chamber access
type UserWithRoomChamberAccess struct {
	User         User                  `json:"user"`
	RoomChambers []RoomChamberResponse `json:"room_chambers"`
}
