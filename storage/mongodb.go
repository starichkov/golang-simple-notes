package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"golang-simple-notes/model"
)

// MongoDBStorage implements NoteStorage using MongoDB
type MongoDBStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMongoDBStorage creates a new instance of MongoDBStorage
func NewMongoDBStorage(uri, dbName, collectionName string) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get collection
	collection := client.Database(dbName).Collection(collectionName)

	// Create indexes for better performance
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &MongoDBStorage{
		client:     client,
		collection: collection,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Close closes any resources used by the storage
func (s *MongoDBStorage) Close() error {
	defer s.cancel()
	return s.client.Disconnect(s.ctx)
}

// Create adds a new note to the storage
func (s *MongoDBStorage) Create(note *model.Note) error {
	_, err := s.collection.InsertOne(s.ctx, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// Get retrieves a note by its ID
func (s *MongoDBStorage) Get(id string) (*model.Note, error) {
	var note model.Note
	err := s.collection.FindOne(s.ctx, bson.M{"id": id}).Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	return &note, nil
}

// GetAll retrieves all notes from the storage
func (s *MongoDBStorage) GetAll() ([]*model.Note, error) {
	cursor, err := s.collection.Find(s.ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all notes: %w", err)
	}
	defer cursor.Close(s.ctx)

	var notes []*model.Note
	if err := cursor.All(s.ctx, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode notes: %w", err)
	}
	return notes, nil
}

// Update updates an existing note
func (s *MongoDBStorage) Update(note *model.Note) error {
	// Update the note with the current time
	note.UpdatedAt = time.Now()

	result, err := s.collection.ReplaceOne(s.ctx, bson.M{"id": note.ID}, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrNoteNotFound
	}

	return nil
}

// Delete removes a note from the storage
func (s *MongoDBStorage) Delete(id string) error {
	result, err := s.collection.DeleteOne(s.ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrNoteNotFound
	}

	return nil
}
