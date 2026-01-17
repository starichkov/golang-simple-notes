# Development Guidelines for Notes API

This document provides essential information for developers working on the Notes API project.

## 🛠 Build and Configuration

### Prerequisites
- Go 1.25.6 or later (as specified in `go.mod`).
- Docker and Docker Compose (for running with databases or integration tests).

### Compilation
To compile the application with reduced binary size:
```bash
go mod tidy
go build -ldflags="-s -w"
```
- `-s`: removes the symbol table and debug info.
- `-w`: removes DWARF debugging information.

### Configuration
The application is configured via environment variables:

| Variable             | Description                                        | Default                     |
|----------------------|----------------------------------------------------|-----------------------------|
| `STORAGE_TYPE`       | Type of storage: `memory`, `couchdb`, or `mongodb` | `memory`                    |
| `COUCHDB_URL`        | URL of the CouchDB server                          | `http://localhost:5984`     |
| `COUCHDB_DB`         | Name of the CouchDB database                       | `notes`                     |
| `MONGODB_URI`        | URI of the MongoDB server                          | `mongodb://localhost:27017` |
| `MONGODB_DB`         | Name of the MongoDB database                       | `notes`                     |
| `MONGODB_COLLECTION` | Name of the MongoDB collection                     | `notes`                     |

## 🧪 Testing

### Running Tests
- **All tests**: `go test ./...`
- **Unit tests only**: `go test -short ./...` (skips integration tests that require Docker).
- **Specific package**: `go test ./storage`

### Integration Tests
Integration tests for CouchDB and MongoDB use `testcontainers-go`. They are automatically skipped in `short` mode.

### Adding New Tests
When adding new tests, follow these conventions:
1. **Mocks**: Use package-local mock implementations for storage if you want to test handlers or services in isolation.
2. **Integration**: If you need an actual database, use the shared container setup if available in the package (see `TestMain` in `main_test.go` or `storage/shared_test.go`).
3. **Short Mode**: Always wrap integration tests with `if testing.Short() { t.Skip(...) }`.

### Demonstration Test
Here is a simple test example demonstrating how to test the storage logic using the in-memory implementation:

```go
func TestSimpleStorage(t *testing.T) {
    ctx := context.Background()
    store := storage.NewInMemoryStorage()

    note := &model.Note{
        ID:      "test-1",
        Title:   "Test Note",
        Content: "Hello World",
    }

    // Create
    if err := store.Create(ctx, note); err != nil {
        t.Fatalf("Create failed: %v", err)
    }

    // Get
    retrieved, err := store.Get(ctx, "test-1")
    if err != nil || retrieved.Title != "Test Note" {
        t.Errorf("Get failed or returned wrong data")
    }
}
```

## 📝 Additional Information

### Code Style
- Follow standard Go idioms and `gofmt`.
- Use `context.Context` for all IO-bound operations and propagate it through the call stack.
- Errors should be handled explicitly. Package-specific error variables (like `storage.ErrNoteNotFound`) are used for consistent error checking.

### Architecture
The project follows a clean architecture:
- `model`: Domain entities.
- `storage`: Data access layer with multiple implementations.
- `rest`: HTTP handlers.
- `grpc`: gRPC service implementation.
- `proto`: gRPC service definitions.

### Shared Test State
The project uses a file-based locking mechanism (`testcontainers-shared-state.lock` in `os.TempDir()`) to share test containers across different test runs/packages to speed up integration testing.
