package main

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Test with default values
	config := NewConfig()

	if config.StorageType != "memory" {
		t.Errorf("Expected StorageType to be 'memory', got %s", config.StorageType)
	}
	if config.CouchDBURL != "http://localhost:5984" {
		t.Errorf("Expected CouchDBURL to be 'http://localhost:5984', got %s", config.CouchDBURL)
	}
	if config.CouchDBName != "notes" {
		t.Errorf("Expected CouchDBName to be 'notes', got %s", config.CouchDBName)
	}
	if config.MongoDBURI != "mongodb://localhost:27017" {
		t.Errorf("Expected MongoDBURI to be 'mongodb://localhost:27017', got %s", config.MongoDBURI)
	}
	if config.MongoDBName != "notes" {
		t.Errorf("Expected MongoDBName to be 'notes', got %s", config.MongoDBName)
	}
	if config.MongoDBCollection != "notes" {
		t.Errorf("Expected MongoDBCollection to be 'notes', got %s", config.MongoDBCollection)
	}
	if config.RESTPort != ":8080" {
		t.Errorf("Expected RESTPort to be ':8080', got %s", config.RESTPort)
	}
	if config.GRPCPort != ":8081" {
		t.Errorf("Expected GRPCPort to be ':8081', got %s", config.GRPCPort)
	}

	// Test environment variable override
	t.Setenv("STORAGE_TYPE", "couchdb")
	t.Setenv("COUCHDB_URL", "http://test:5984")
	t.Setenv("COUCHDB_DB", "testdb")
	t.Setenv("MONGODB_URI", "mongodb://test:27017")
	t.Setenv("MONGODB_DB", "testdb")
	t.Setenv("MONGODB_COLLECTION", "testcoll")

	config = NewConfig()
	if config.StorageType != "couchdb" {
		t.Errorf("Expected StorageType to be 'couchdb', got %s", config.StorageType)
	}
	if config.CouchDBURL != "http://test:5984" {
		t.Errorf("Expected CouchDBURL to be 'http://test:5984', got %s", config.CouchDBURL)
	}
	if config.CouchDBName != "testdb" {
		t.Errorf("Expected CouchDBName to be 'testdb', got %s", config.CouchDBName)
	}
	if config.MongoDBURI != "mongodb://test:27017" {
		t.Errorf("Expected MongoDBURI to be 'mongodb://test:27017', got %s", config.MongoDBURI)
	}
	if config.MongoDBName != "testdb" {
		t.Errorf("Expected MongoDBName to be 'testdb', got %s", config.MongoDBName)
	}
	if config.MongoDBCollection != "testcoll" {
		t.Errorf("Expected MongoDBCollection to be 'testcoll', got %s", config.MongoDBCollection)
	}

}

func TestGetEnv(t *testing.T) {
	// Test default value when environment variable is not set
	value := getEnv("NONEXISTENT_VAR", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got %s", value)
	}

	// Test environment variable override
	t.Setenv("TEST_VAR", "test_value")
	value = getEnv("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %s", value)
	}
}
