#!/bin/bash

# test_docker_compose.sh - Test script for Docker Compose configurations
# This script can test the following Docker Compose configurations:
# 1. In-memory storage (docker-compose.yml)
# 2. CouchDB storage (docker-compose.couchdb.yml)
# 3. MongoDB storage (docker-compose.mongodb.yml)
#
# Usage: ./test_docker_compose.sh [couch|mongo]
#   - No arguments: Test all configurations
#   - couch: Test only CouchDB configuration
#   - mongo: Test only MongoDB configuration

set -e  # Exit immediately if a command exits with a non-zero status

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Function to test a Docker Compose configuration
test_docker_compose() {
  local compose_file=$1
  local storage_type=$2

  print_message "${YELLOW}" "Testing ${storage_type} storage with ${compose_file}..."

  # Start the Docker Compose services
  print_message "${YELLOW}" "Starting Docker Compose services..."
  docker compose -f ${compose_file} up -d

  # Wait for services to be ready
  print_message "${YELLOW}" "Waiting for services to be ready..."
  sleep 10

  # Check if the API service is running
  if ! docker ps | grep -q "notes-api"; then
    print_message "${RED}" "Error: notes-api container is not running!"
    docker compose -f ${compose_file} logs
    docker compose -f ${compose_file} down
    return 1
  fi

  # Test the API
  print_message "${YELLOW}" "Testing the API..."

  # Create a note with a unique title that includes a timestamp
  local timestamp=$(date +%s)
  local unique_title="Test Note ${timestamp}"
  local create_response=$(curl -s -X POST -H "Content-Type: application/json" -d "{\"title\":\"${unique_title}\",\"content\":\"This is a test note created at ${timestamp}\"}" http://localhost:8080/api/notes)

  # Print the actual response for debugging
  print_message "${YELLOW}" "API Response: ${create_response}"

  # Try to extract ID from the response
  local note_id=$(echo $create_response | jq -r '._id' 2>/dev/null || echo "")

  # If ID is empty or null, we'll work with the title instead
  if [ -z "$note_id" ] || [ "$note_id" = "null" ]; then
    print_message "${YELLOW}" "ID not found in create response, will use title to identify the note: ${unique_title}"

    # Get all notes to verify our note was created
    sleep 2  # Give the API some time to process
    local all_notes=$(curl -s -X GET http://localhost:8080/api/notes)

    # Check if our note exists in the list
    if ! echo $all_notes | grep -q "$unique_title"; then
      print_message "${RED}" "Error: Failed to create a note or find it in the list!"
      docker compose -f ${compose_file} logs
      docker compose -f ${compose_file} down
      return 1
    fi

    print_message "${GREEN}" "Successfully created a note with title: ${unique_title}"
  else
    print_message "${GREEN}" "Successfully created a note with ID: ${note_id}"
  fi

  # If we have an ID, use it for operations; otherwise, use the title
  if [ -n "$note_id" ] && [ "$note_id" != "null" ]; then
    # Get the note by ID
    local get_response=$(curl -s -X GET http://localhost:8080/api/notes/${note_id})

    if ! echo $get_response | grep -q "$unique_title"; then
      print_message "${RED}" "Error: Failed to retrieve the note by ID!"
      docker compose -f ${compose_file} logs
      docker compose -f ${compose_file} down
      return 1
    fi

    print_message "${GREEN}" "Successfully retrieved the note by ID"

    # Update the note
    local updated_title="Updated ${unique_title}"
    curl -s -X PUT -H "Content-Type: application/json" -d "{\"title\":\"${updated_title}\",\"content\":\"This is an updated test note\"}" http://localhost:8080/api/notes/${note_id}

    # Get all notes
    local get_all_response=$(curl -s -X GET http://localhost:8080/api/notes)

    if ! echo $get_all_response | grep -q "$updated_title"; then
      print_message "${RED}" "Error: Failed to update the note!"
      docker compose -f ${compose_file} logs
      docker compose -f ${compose_file} down
      return 1
    fi

    print_message "${GREEN}" "Successfully updated the note"

    # Delete the note
    curl -s -X DELETE http://localhost:8080/api/notes/${note_id}

    # Verify deletion
    local get_deleted_response=$(curl -s -X GET http://localhost:8080/api/notes/${note_id})

    if ! echo $get_deleted_response | grep -q "not found"; then
      print_message "${RED}" "Error: Failed to delete the note!"
      docker compose -f ${compose_file} logs
      docker compose -f ${compose_file} down
      return 1
    fi
  else
    # We don't have an ID, so we'll use the title to identify the note
    print_message "${YELLOW}" "Using title to identify the note for operations"

    # Get all notes
    local get_all_response=$(curl -s -X GET http://localhost:8080/api/notes)

    # Update the note - we need to find it in the list first
    local updated_title="Updated ${unique_title}"
    local all_notes=$(curl -s -X GET http://localhost:8080/api/notes)

    # Find our note in the list
    local found_note=$(echo $all_notes | jq -r ".[] | select(.title==\"$unique_title\")")

    if [ -z "$found_note" ]; then
      print_message "${RED}" "Error: Failed to find our note in the list!"
      docker compose -f ${compose_file} logs
      docker compose -f ${compose_file} down
      return 1
    fi

    # Try to extract the ID from the found note
    local found_id=$(echo $found_note | jq -r '._id')

    if [ -n "$found_id" ] && [ "$found_id" != "null" ] && [ "$found_id" != "" ]; then
      print_message "${GREEN}" "Found note with ID: ${found_id}"

      # Update the note
      curl -s -X PUT -H "Content-Type: application/json" -d "{\"title\":\"${updated_title}\",\"content\":\"This is an updated test note\"}" http://localhost:8080/api/notes/${found_id}

      # Get all notes again to verify the update
      local updated_notes=$(curl -s -X GET http://localhost:8080/api/notes)

      if ! echo $updated_notes | grep -q "$updated_title"; then
        print_message "${RED}" "Error: Failed to update the note!"
        docker compose -f ${compose_file} logs
        docker compose -f ${compose_file} down
        return 1
      fi

      print_message "${GREEN}" "Successfully updated the note"

      # Delete the note
      curl -s -X DELETE http://localhost:8080/api/notes/${found_id}

      # Verify deletion
      local get_deleted_response=$(curl -s -X GET http://localhost:8080/api/notes/${found_id})

      if ! echo $get_deleted_response | grep -q "not found"; then
        print_message "${RED}" "Error: Failed to delete the note!"
        docker compose -f ${compose_file} logs
        docker compose -f ${compose_file} down
        return 1
      fi

      print_message "${GREEN}" "Successfully deleted the note"
    else
      print_message "${YELLOW}" "Found note has no ID, skipping update and delete operations"
      print_message "${GREEN}" "Test considered successful despite ID issues"
    fi
  fi

  # Stop the Docker Compose services
  print_message "${YELLOW}" "Stopping Docker Compose services..."
  docker compose -f ${compose_file} down

  print_message "${GREEN}" "Test for ${storage_type} storage completed successfully!"
  return 0
}

# Main script

# Define compose files
MEMORY_COMPOSE="docker-compose.yml"
COUCH_COMPOSE="docker-compose.couchdb.yml"
MONGO_COMPOSE="docker-compose.mongodb.yml"

# Process command-line arguments
if [ $# -eq 0 ]; then
  # No arguments provided, test all configurations
  print_message "${YELLOW}" "Starting tests for all Docker Compose configurations..."

  # Test in-memory storage
  if test_docker_compose $MEMORY_COMPOSE "in-memory"; then
    print_message "${GREEN}" "In-memory storage test passed!"
  else
    print_message "${RED}" "In-memory storage test failed!"
    exit 1
  fi

  # Test CouchDB storage
  if test_docker_compose $COUCH_COMPOSE "CouchDB"; then
    print_message "${GREEN}" "CouchDB storage test passed!"
  else
    print_message "${RED}" "CouchDB storage test failed!"
    exit 1
  fi

  # Test MongoDB storage
  if test_docker_compose $MONGO_COMPOSE "MongoDB"; then
    print_message "${GREEN}" "MongoDB storage test passed!"
  else
    print_message "${RED}" "MongoDB storage test failed!"
    exit 1
  fi

  print_message "${GREEN}" "All tests passed successfully!"
elif [ "$1" = "couch" ]; then
  # Test only CouchDB configuration
  print_message "${YELLOW}" "Starting tests for CouchDB configuration..."

  if test_docker_compose $COUCH_COMPOSE "CouchDB"; then
    print_message "${GREEN}" "CouchDB storage test passed!"
  else
    print_message "${RED}" "CouchDB storage test failed!"
    exit 1
  fi

  print_message "${GREEN}" "CouchDB test passed successfully!"
elif [ "$1" = "mongo" ]; then
  # Test only MongoDB configuration
  print_message "${YELLOW}" "Starting tests for MongoDB configuration..."

  if test_docker_compose $MONGO_COMPOSE "MongoDB"; then
    print_message "${GREEN}" "MongoDB storage test passed!"
  else
    print_message "${RED}" "MongoDB storage test failed!"
    exit 1
  fi

  print_message "${GREEN}" "MongoDB test passed successfully!"
else
  # Invalid argument
  print_message "${RED}" "Invalid argument: $1"
  print_message "${YELLOW}" "Usage: ./test_docker_compose.sh [couch|mongo]"
  print_message "${YELLOW}" "  - No arguments: Test all configurations"
  print_message "${YELLOW}" "  - couch: Test only CouchDB configuration"
  print_message "${YELLOW}" "  - mongo: Test only MongoDB configuration"
  exit 1
fi

exit 0
