package main

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Test with default values
	config := NewConfig()

	if config.StorageType != "memory" {
		t.Errorf("Expected StorageType to be 'memory', got '%s'", config.StorageType)
	}

	if config.CouchDBURL != "http://localhost:5984" {
		t.Errorf("Expected CouchDBURL to be 'http://localhost:5984', got '%s'", config.CouchDBURL)
	}

	if config.MongoDBURI != "mongodb://localhost:27017" {
		t.Errorf("Expected MongoDBURI to be 'mongodb://localhost:27017', got '%s'", config.MongoDBURI)
	}

	if config.RESTPort != ":8080" {
		t.Errorf("Expected RESTPort to be ':8080', got '%s'", config.RESTPort)
	}

	if config.GRPCPort != ":8081" {
		t.Errorf("Expected GRPCPort to be ':8081', got '%s'", config.GRPCPort)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := getEnv("TEST_VAR", "default_value")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test with environment variable not set
	value = getEnv("NON_EXISTENT_VAR", "default_value")
	if value != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", value)
	}
}
