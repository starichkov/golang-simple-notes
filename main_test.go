package main

import (
	"context"
	"fmt"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
