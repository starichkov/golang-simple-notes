package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang-simple-notes/model"
)

// MongoDBStorage implements NoteStorage using MongoDB
type MongoDBStorage struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// NewMongoDBStorage creates a new MongoDB storage instance
func NewMongoDBStorage(uri, dbName, collectionName string) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoDBStorage{
		client:     client,
		database:   client.Database(dbName),
		collection: client.Database(dbName).Collection(collectionName),
	}, nil
}

// Create adds a new note to MongoDB
func (s *MongoDBStorage) Create(ctx context.Context, note *model.Note) error {
	_, err := s.collection.InsertOne(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to insert note: %w", err)
	}
	return nil
}

// Get retrieves a note by its ID from MongoDB
func (s *MongoDBStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	var note model.Note
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to find note: %w", err)
	}
	return &note, nil
}

// GetAll retrieves all notes from MongoDB
func (s *MongoDBStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find notes: %w", err)
	}
	defer cursor.Close(ctx)

	var notes []*model.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode notes: %w", err)
	}
	return notes, nil
}

// Update updates an existing note in MongoDB
func (s *MongoDBStorage) Update(ctx context.Context, note *model.Note) error {
	result, err := s.collection.ReplaceOne(ctx, bson.M{"_id": note.ID}, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrNoteNotFound
	}
	return nil
}

// Delete removes a note from MongoDB
func (s *MongoDBStorage) Delete(ctx context.Context, id string) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNoteNotFound
	}
	return nil
}

// Close closes the MongoDB connection
func (s *MongoDBStorage) Close(ctx context.Context) error {
	if err := s.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}
