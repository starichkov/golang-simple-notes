package main

import "os"

// Config holds all configuration for the application
type Config struct {
	StorageType       string
	CouchDBURL        string
	CouchDBName       string
	MongoDBURI        string
	MongoDBName       string
	MongoDBCollection string
	RESTPort          string
	GRPCPort          string
}

// NewConfig creates a new Config instance with values from environment variables
func NewConfig() *Config {
	return &Config{
		StorageType:       getEnv("STORAGE_TYPE", "memory"),
		CouchDBURL:        getEnv("COUCHDB_URL", "http://localhost:5984"),
		CouchDBName:       getEnv("COUCHDB_DB", "notes"),
		MongoDBURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:       getEnv("MONGODB_DB", "notes"),
		MongoDBCollection: getEnv("MONGODB_COLLECTION", "notes"),
		RESTPort:          ":8080",
		GRPCPort:          ":8081",
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
