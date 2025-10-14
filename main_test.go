package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	// Shared container instances for all tests in the main package
	sharedMongoContainer    testcontainers.Container
	sharedMongoURI          string
	sharedCouchContainer    testcontainers.Container
	sharedCouchURL          string
	shouldCleanupContainers bool
)

// containerState holds the connection details for shared containers
type containerState struct {
	MongoURI string `json:"mongo_uri"`
	CouchURL string `json:"couch_url"`
	OwnerPID int    `json:"owner_pid"`
}

// getStateFilePath returns the path to the shared state file
func getStateFilePath() string {
	return filepath.Join(os.TempDir(), "testcontainers-shared-state.json")
}

// getLockFilePath returns the path to the lock file
func getLockFilePath() string {
	return filepath.Join(os.TempDir(), "testcontainers-shared-state.lock")
}

// acquireLock attempts to acquire a file-based lock
func acquireLock() (*os.File, error) {
	lockFile := getLockFilePath()
	// Try to create lock file exclusively (wait up to ~2 minutes)
	for i := 0; i < 1200; i++ { // 1200 * 100ms = 120s
		f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// Write our PID to the lock file
			fmt.Fprintf(f, "%d", os.Getpid())
			return f, nil
		}
		// Lock file exists, wait and retry
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("failed to acquire lock after 120 seconds")
}

// releaseLock releases the file-based lock
func releaseLock(f *os.File) {
	if f != nil {
		f.Close()
		os.Remove(getLockFilePath())
	}
}

// loadContainerState loads the container state from file
func loadContainerState() (*containerState, error) {
	stateFile := getStateFilePath()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state containerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// saveContainerState saves the container state to file
func saveContainerState(state *containerState) error {
	stateFile := getStateFilePath()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// TestMain sets up shared test containers for the main package
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Acquire lock to prevent race condition with other packages
	lock, err := acquireLock()
	if err != nil {
		log.Printf("Warning: Failed to acquire lock: %v. Waiting for shared state...", err)
		// If we cannot get the lock, another package is likely starting containers.
		// Wait for the shared state file to appear and reuse those containers.
		deadline := time.Now().Add(2 * time.Minute)
		for time.Now().Before(deadline) {
			if state, err := loadContainerState(); err == nil && state != nil {
				log.Printf("Reusing existing MongoDB container at %s", state.MongoURI)
				log.Printf("Reusing existing CouchDB container at %s", state.CouchURL)
				sharedMongoURI = state.MongoURI
				sharedCouchURL = state.CouchURL
				shouldCleanupContainers = false
				// Run tests without owning the containers
				code := m.Run()
				os.Exit(code)
			}
			time.Sleep(250 * time.Millisecond)
		}
		log.Printf("Warning: Shared state not found after waiting. Integration tests may be skipped.")
		// Proceed without containers (integration tests will self-skip if unavailable)
		code := m.Run()
		os.Exit(code)
	}

	// Try to load existing container state
	state, err := loadContainerState()
	if err == nil && state != nil {
		// Reuse existing containers from another package
		log.Printf("Reusing existing MongoDB container at %s", state.MongoURI)
		log.Printf("Reusing existing CouchDB container at %s", state.CouchURL)
		sharedMongoURI = state.MongoURI
		sharedCouchURL = state.CouchURL
		shouldCleanupContainers = false
		releaseLock(lock)
	} else {
		// Start new containers and save state
		shouldCleanupContainers = true

		// Start shared MongoDB container
		if err := startSharedMongoDBContainer(ctx); err != nil {
			log.Printf("Warning: Failed to start shared MongoDB container: %v. MongoDB tests may fail.", err)
		}

		// Start shared CouchDB container
		if err := startSharedCouchDBContainer(ctx); err != nil {
			log.Printf("Warning: Failed to start shared CouchDB container: %v. CouchDB tests may fail.", err)
		}

		// Save container state for other packages to reuse
		if sharedMongoURI != "" && sharedCouchURL != "" {
			state := &containerState{
				MongoURI: sharedMongoURI,
				CouchURL: sharedCouchURL,
				OwnerPID: os.Getpid(),
			}
			if err := saveContainerState(state); err != nil {
				log.Printf("Warning: Failed to save container state: %v", err)
			}
		}

		releaseLock(lock)
	}

	// Run tests
	code := m.Run()

	// Cleanup containers only if we started them
	if shouldCleanupContainers {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if sharedMongoContainer != nil {
			if err := sharedMongoContainer.Terminate(cleanupCtx); err != nil {
				log.Printf("Failed to terminate MongoDB container: %v", err)
			}
		}

		if sharedCouchContainer != nil {
			if err := sharedCouchContainer.Terminate(cleanupCtx); err != nil {
				log.Printf("Failed to terminate CouchDB container: %v", err)
			}
		}

		// Clean up state file
		os.Remove(getStateFilePath())
	}

	os.Exit(code)
}

func startSharedMongoDBContainer(ctx context.Context) error {
	container, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:7.0.25-jammy"))
	if err != nil {
		return fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	sharedMongoContainer = container

	// Get connection string
	mongoURI, err := container.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MongoDB connection string: %w", err)
	}

	sharedMongoURI = mongoURI

	log.Printf("Started shared MongoDB container at %s", sharedMongoURI)
	return nil
}

func startSharedCouchDBContainer(ctx context.Context) error {
	req := testcontainers.ContainerRequest{
		Image:        "couchdb:3.4.3",
		ExposedPorts: []string{"5984/tcp"},
		WaitingFor:   wait.ForListeningPort("5984/tcp"),
		Env: map[string]string{
			"COUCHDB_USER":     "admin",
			"COUCHDB_PASSWORD": "password",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to start CouchDB container: %w", err)
	}

	sharedCouchContainer = container

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get CouchDB container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5984")
	if err != nil {
		return fmt.Errorf("failed to get CouchDB container port: %w", err)
	}

	sharedCouchURL = fmt.Sprintf("http://admin:password@%s:%s", host, port.Port())

	log.Printf("Started shared CouchDB container at %s", sharedCouchURL)
	return nil
}

// MockStorage is a simple implementation of NoteStorage for testing
type MockStorage struct {
	notes []*model.Note
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		notes: make([]*model.Note, 0),
	}
}

func (s *MockStorage) Create(ctx context.Context, note *model.Note) error {
	s.notes = append(s.notes, note)
	return nil
}

func (s *MockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	for _, note := range s.notes {
		if note.ID == id {
			return note, nil
		}
	}
	return nil, storage.ErrNoteNotFound
}

func (s *MockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	return s.notes, nil
}

func (s *MockStorage) Update(ctx context.Context, note *model.Note) error {
	for i, n := range s.notes {
		if n.ID == note.ID {
			s.notes[i] = note
			return nil
		}
	}
	return storage.ErrNoteNotFound
}

func (s *MockStorage) Delete(ctx context.Context, id string) error {
	for i, note := range s.notes {
		if note.ID == id {
			s.notes = append(s.notes[:i], s.notes[i+1:]...)
			return nil
		}
	}
	return storage.ErrNoteNotFound
}

func (s *MockStorage) Close(ctx context.Context) error {
	return nil
}

// ErrorMockStorage is a mock storage that returns an error on Create
type ErrorMockStorage struct{}

func (s *ErrorMockStorage) Create(ctx context.Context, note *model.Note) error {
	return fmt.Errorf("mock error")
}

func (s *ErrorMockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	return nil, storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	return nil, nil
}

func (s *ErrorMockStorage) Update(ctx context.Context, note *model.Note) error {
	return storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) Delete(ctx context.Context, id string) error {
	return storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) Close(ctx context.Context) error {
	return nil
}

type MyLogConsumer struct{}

func (c *MyLogConsumer) Accept(log testcontainers.Log) {
	fmt.Printf("Log: %s\n", string(log.Content))
}

// startCouchDBContainer starts a CouchDB container and returns the container and connection URL
func startCouchDBContainer(ctx context.Context) (testcontainers.Container, string, error) {
	consumer := &MyLogConsumer{}
	req := testcontainers.ContainerRequest{
		Image:        "couchdb:3.4.3",
		ExposedPorts: []string{"5984/tcp"},
		WaitingFor:   wait.ForListeningPort("5984/tcp"),
		Env: map[string]string{
			"COUCHDB_USER":     "admin",
			"COUCHDB_PASSWORD": "password",
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{
				consumer,
			},
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}
	port, err := container.MappedPort(ctx, "5984")
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}
	url := fmt.Sprintf("http://admin:password@%s:%s", host, port.Port())
	return container, url, nil
}
