package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the database connection
type MongoDB struct {
	Client                      *mongo.Client
	Database                    *mongo.Database
	ChambersCollection          *mongo.Collection
	ExperimentsCollection       *mongo.Collection
	UsersCollection             *mongo.Collection
	SessionsCollection          *mongo.Collection
	APITokensCollection         *mongo.Collection
	UserChamberAccessCollection *mongo.Collection
}

// Connect establishes a connection to MongoDB
func Connect(uri, databaseName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// Get database and collections
	db := client.Database(databaseName)

	return &MongoDB{
		Client:                      client,
		Database:                    db,
		ChambersCollection:          db.Collection("chambers"),
		ExperimentsCollection:       db.Collection("experiments"),
		UsersCollection:             db.Collection("users"),
		SessionsCollection:          db.Collection("sessions"),
		APITokensCollection:         db.Collection("api_tokens"),
		UserChamberAccessCollection: db.Collection("user_chamber_access"),
	}, nil
}

// Disconnect closes the database connection
func (m *MongoDB) Disconnect(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
