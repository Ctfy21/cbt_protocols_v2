package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the MongoDB client and database references
type MongoDB struct {
	Client                *mongo.Client
	Database              *mongo.Database
	ChambersCollection    *mongo.Collection
	ExperimentsCollection *mongo.Collection
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(10).
		SetMinPoolSize(1).
		SetMaxConnIdleTime(30 * time.Second)

	// Connect to MongoDB
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Test the connection
	log.Println("Testing MongoDB connection...")
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// Get database and collections
	database := client.Database(dbName)

	db := &MongoDB{
		Client:                client,
		Database:              database,
		ChambersCollection:    database.Collection("chambers"),
		ExperimentsCollection: database.Collection("experiments"),
	}

	log.Printf("✅ Successfully connected to MongoDB database: %s", dbName)
	return db, nil
}

// Disconnect gracefully disconnects from MongoDB
func (db *MongoDB) Disconnect(ctx context.Context) error {
	log.Println("Disconnecting from MongoDB...")
	err := db.Client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %v", err)
	}
	log.Println("✅ Successfully disconnected from MongoDB")
	return nil
}

// HealthCheck checks if the database connection is healthy
func (db *MongoDB) HealthCheck(ctx context.Context) error {
	return db.Client.Ping(ctx, nil)
}
