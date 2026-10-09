package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB encapsulates MongoDB client and collection references
type DB struct {
	Client        *mongo.Client
	Database      *mongo.Database
	Members       *mongo.Collection
	Groups        *mongo.Collection
	GlobalUsers   *mongo.Collection
}

// Connect establishes a connection to MongoDB and sets up indexes
func Connect(uri, dbName string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("[DATABASE] Successfully connected to MongoDB (%s)", dbName)

	database := client.Database(dbName)
	db := &DB{
		Client:      client,
		Database:    database,
		Members:     database.Collection("members"),
		Groups:      database.Collection("groups"),
		GlobalUsers: database.Collection("global_users"),
	}

	// Ensure Indexes
	db.ensureIndexes()

	return db, nil
}

func (db *DB) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Members: Compound Unique Index (ChatID + UserID)
	_, _ = db.Members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "chat_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// 2. Members: Sorting indexes for leaderboards
	_, _ = db.Members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}, {Key: "daily_count", Value: -1}},
	})
	_, _ = db.Members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}, {Key: "weekly_count", Value: -1}},
	})
	_, _ = db.Members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}, {Key: "monthly_count", Value: -1}},
	})
	_, _ = db.Members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}, {Key: "all_time_count", Value: -1}},
	})
	_, _ = db.Members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}, {Key: "xp", Value: -1}},
	})

	// 3. Groups: Unique ChatID
	_, _ = db.Groups.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "chat_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// 4. GlobalUsers: Unique UserID & XP index
	_, _ = db.GlobalUsers.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	_, _ = db.GlobalUsers.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "global_xp", Value: -1}},
	})

	log.Printf("[DATABASE] Indexes verified successfully.")
}

// Disconnect cleans up database connection
func (db *DB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = db.Client.Disconnect(ctx)
}
