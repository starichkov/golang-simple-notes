// This file contains the MongoDB implementation of the NoteStorage interface.
// It uses the official MongoDB Go driver to interact with MongoDB.
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

// MongoDBStorage implements NoteStorage using MongoDB.
// MongoDB is a document-oriented NoSQL database that stores data as BSON documents.
// It's designed for scalability and performance, making it suitable for applications
// that need to handle large volumes of data.
type MongoDBStorage struct {
	client     *mongo.Client     // MongoDB client for connecting to the server
	database   *mongo.Database   // Database handle
	collection *mongo.Collection // Collection handle for storing notes
}

// NewMongoDBStorage creates a new MongoDB storage instance.
// It connects to the MongoDB server at the specified URI, and uses the specified
// database and collection for storing notes.
//
// Parameters:
//   - uri: The MongoDB connection string (e.g., "mongodb://admin:password@localhost:27017")
//   - dbName: The name of the database to use
//   - collectionName: The name of the collection to store notes in
//
// Returns:
//   - A pointer to a new MongoDBStorage instance
//   - An error if the connection fails
func NewMongoDBStorage(uri, dbName, collectionName string) (*MongoDBStorage, error) {
	// Create a context with a timeout for the connection (configurable via env)
	mongoTimeoutMs := getenvInt("MONGODB_CONNECT_TIMEOUT_MS", 10000)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoTimeoutMs)*time.Millisecond)
	defer cancel() // Ensure the context is canceled when the function returns

	// Connect to MongoDB using the provided URI
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify the connection is working
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Return a new MongoDBStorage instance with the client, database, and collection
	return &MongoDBStorage{
		client:     client,
		database:   client.Database(dbName),
		collection: client.Database(dbName).Collection(collectionName),
	}, nil
}

// Create adds a new note to MongoDB.
// It uses the MongoDB driver's InsertOne method to store the note as a BSON document.
// The note's ID is used as the document ID in MongoDB.
func (s *MongoDBStorage) Create(ctx context.Context, note *model.Note) error {
	// Insert the note into MongoDB
	// MongoDB will automatically convert the Go struct to BSON format
	_, err := s.collection.InsertOne(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to insert note: %w", err)
	}
	return nil
}

// Get retrieves a note by its ID from MongoDB.
// It returns the note if found, or ErrNoteNotFound if no note with the specified ID exists.
func (s *MongoDBStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	var note model.Note

	// Find a document with the specified ID and decode it into the note struct
	// bson.M is a map-like type for representing BSON documents
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&note)
	if err != nil {
		// Check if the error is "document not found"
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to find note: %w", err)
	}
	return &note, nil
}

// GetAll retrieves all notes from MongoDB.
// It returns a slice of all notes in the collection, which may be empty if there are no notes.
func (s *MongoDBStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	// Find all documents in the collection
	// An empty bson.M{} filter matches all documents
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find notes: %w", err)
	}
	// Ensure the cursor is closed when the function returns
	defer cursor.Close(ctx)

	// Decode all documents into a slice of Note pointers
	var notes []*model.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode notes: %w", err)
	}
	return notes, nil
}

// Update updates an existing note in MongoDB.
// It returns ErrNoteNotFound if no note with the specified ID exists.
func (s *MongoDBStorage) Update(ctx context.Context, note *model.Note) error {
	// Replace the entire document with the new note
	// ReplaceOne is used instead of UpdateOne to ensure all fields are updated
	result, err := s.collection.ReplaceOne(ctx, bson.M{"_id": note.ID}, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	// Check if any document was matched (updated)
	// If no document was matched, it means the note doesn't exist
	if result.MatchedCount == 0 {
		return ErrNoteNotFound
	}
	return nil
}

// Delete removes a note from MongoDB.
// It returns ErrNoteNotFound if no note with the specified ID exists.
func (s *MongoDBStorage) Delete(ctx context.Context, id string) error {
	// Delete the document with the specified ID
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	// Check if any document was deleted
	// If no document was deleted, it means the note doesn't exist
	if result.DeletedCount == 0 {
		return ErrNoteNotFound
	}
	return nil
}

// Close closes the MongoDB connection.
// This should be called when the application is shutting down to release resources.
func (s *MongoDBStorage) Close(ctx context.Context) error {
	// Disconnect from MongoDB
	if err := s.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}
