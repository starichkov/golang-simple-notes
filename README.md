[![Author](https://img.shields.io/badge/Author-Vadim%20Starichkov-blue?style=for-the-badge)](https://github.com/starichkov)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/starichkov/golang-simple-notes/build.yml?style=for-the-badge)](https://github.com/starichkov/golang-simple-notes/actions/workflows/build.yml)
[![Codecov](https://img.shields.io/codecov/c/github/starichkov/golang-simple-notes?style=for-the-badge)](https://codecov.io/gh/starichkov/golang-simple-notes)
[![GitHub License](https://img.shields.io/github/license/starichkov/golang-simple-notes?style=for-the-badge)](https://github.com/starichkov/golang-simple-notes/blob/main/LICENSE.md)

Notes API: Golang with NoSQL Databases
=

A simple microservice for notes management with REST and gRPC APIs, with support for multiple storage backends.

**This project is generated using JetBrains Junie and several other AI coding agents, to evaluate agents capabilities.**

## 👨‍💻 Author

**Vadim Starichkov** | [GitHub](https://github.com/starichkov) | [LinkedIn](https://www.linkedin.com/in/vadim-starichkov/)

*Developed to demonstrate modern Golang development practices and patterns*

## Features

- REST API for CRUD operations on notes
- gRPC API for CRUD operations on notes
- Multiple storage backends:
  - In-memory storage (for development/testing)
  - CouchDB storage
  - MongoDB storage
- Storage-agnostic design
- Docker and Docker Compose support for easy deployment

## Prerequisites

- Docker and Docker Compose
- Go 1.24.5 or later (for local development)

## Compilation

To compile the application, you can use the following command:

```bash
# helps to keep unnecessary dependencies out, which can indirectly reduce size of the binary
go mod tidy
go build -ldflags="-s -w"
```

where:
- removes the symbol table and debug info. `-s`
- removes DWARF debugging information. `-w`

## Running with Docker Compose

The easiest way to run the application is using Docker Compose, which will start both the API services and the databases:

```bash
docker-compose up -d
```

This will:
1. Start a CouchDB instance with authentication
2. Start a MongoDB instance with authentication
3. Build and start two Notes API services:
   - One connected to CouchDB
   - One connected to MongoDB

The services will be available at:

### CouchDB Version
- REST API: http://localhost:8080/api/notes
- gRPC API: localhost:8081
- CouchDB: http://localhost:5984 (admin:password)

### MongoDB Version
- REST API: http://localhost:8082/api/notes
- gRPC API: localhost:8083
- MongoDB: mongodb://localhost:27017 (admin:password)

You can run just one of the services if you prefer:

```bash
# Run only the CouchDB version
docker-compose up -d couchdb notes-api-couchdb

# Run only the MongoDB version
docker-compose up -d mongodb notes-api-mongodb
```

To stop the services:

```bash
docker-compose down
```

To stop the services and remove the volumes:

```bash
docker-compose down -v
```

## Running Locally

You can run the application locally with different storage backends:

### Using In-Memory Storage

The simplest way to run the application is with in-memory storage (no database required):

```bash
export STORAGE_TYPE=memory
go run main.go
```

### Using CouchDB

To use CouchDB, you need to have CouchDB running. You can start CouchDB using Docker:

```bash
docker run -d -p 5984:5984 --name couchdb \
  -e COUCHDB_USER=admin \
  -e COUCHDB_PASSWORD=password \
  couchdb:3.4.3
```

Then, set the environment variables and run the application:

```bash
export STORAGE_TYPE=couchdb
export COUCHDB_URL=http://admin:password@localhost:5984
export COUCHDB_DB=notes
go run main.go
```

### Using MongoDB

To use MongoDB, you need to have MongoDB running. You can start MongoDB using Docker:

```bash
docker run -d -p 27017:27017 --name mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:7.0.20-jammy
```

Then, set the environment variables and run the application:

```bash
export STORAGE_TYPE=mongodb
export MONGODB_URI=mongodb://admin:password@localhost:27017
export MONGODB_DB=notes
export MONGODB_COLLECTION=notes
go run main.go
```

## API Endpoints

### REST API

- `GET /api/notes` - List all notes
- `GET /api/notes/{id}` - Get a note by ID
- `POST /api/notes` - Create a new note
- `PUT /api/notes/{id}` - Update a note
- `DELETE /api/notes/{id}` - Delete a note

### Example Requests

#### Create a Note

```bash
curl -X POST http://localhost:8080/api/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"My Note","content":"This is the content of my note"}'
```

#### Get All Notes

```bash
curl http://localhost:8080/api/notes
```

#### Get a Note by ID

```bash
curl http://localhost:8080/api/notes/{id}
```

#### Update a Note

```bash
curl -X PUT http://localhost:8080/api/notes/{id} \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","content":"Updated content"}'
```

#### Delete a Note

```bash
curl -X DELETE http://localhost:8080/api/notes/{id}
```

## Architecture

The application follows a clean architecture approach with a storage-agnostic design:

- `model` - Contains the domain model (Note)
- `storage` - Contains the storage interface and implementations:
  - `storage.go` - Defines the NoteStorage interface and in-memory implementation
  - `couchdb.go` - CouchDB implementation of the NoteStorage interface using the Kivik library
  - `mongodb.go` - MongoDB implementation of the NoteStorage interface
- `rest` - Contains the REST API handlers
- `grpc` - Contains the gRPC service implementation
- `proto` - Contains the Protocol Buffers definitions

The storage-agnostic design allows the application to work with different storage backends without changing the core business logic. The NoteStorage interface abstracts away the details of how notes are stored and retrieved, making it easy to add new storage implementations.

## Testing

The project includes comprehensive tests for all components:

### Running Tests

To run all tests:

```bash
go test ./...
```

To run tests for a specific package:

```bash
go test ./model
go test ./storage
go test ./rest
go test ./grpc
```

To run tests without external dependencies (skipping integration tests):

```bash
go test -short ./...
```

### Test Coverage

The tests cover:

- **Model Tests**: Tests for the Note model and ID generation
- **Storage Tests**: 
  - In-memory storage implementation
  - MongoDB storage implementation (integration tests, requires MongoDB)
  - CouchDB storage implementation (integration tests, requires CouchDB)
- **REST API Tests**: Tests for all REST endpoints using a mock storage
- **gRPC Service Tests**: Tests for all gRPC service methods using a mock storage

The integration tests for MongoDB and CouchDB are skipped when running with the `-short` flag, as they require actual database instances.

### Continuous Integration

This project uses GitHub Actions for continuous integration. The workflow includes:

- Building the application
- Running unit tests
- Running integration tests with code coverage reporting
- Building the Docker image
- Testing the application with Docker Compose

The workflow is defined in `.github/workflows/build.yml` and runs automatically on pushes to the main branch and pull requests.

To view the latest build status and test results, click on the build status badge at the top of this README or visit the [Actions tab](https://github.com/starichkov/golang-simple-notes/actions) in the GitHub repository.

#### Code Coverage

The GitHub Actions workflow generates code coverage reports for the tests. These reports are available as artifacts in the workflow runs and can be downloaded to analyze test coverage.

Additionally, code coverage reports are automatically uploaded to [Codecov](https://codecov.io) for visualization and tracking of coverage trends over time.

## Configuration

The application can be configured using environment variables:

### General Configuration
- `STORAGE_TYPE` - The type of storage to use: "memory", "couchdb", or "mongodb" (default: "memory")

### CouchDB Configuration
- `COUCHDB_URL` - The URL of the CouchDB server (default: "http://localhost:5984")
- `COUCHDB_DB` - The name of the CouchDB database (default: "notes")

### MongoDB Configuration
- `MONGODB_URI` - The URI of the MongoDB server (default: "mongodb://localhost:27017")
- `MONGODB_DB` - The name of the MongoDB database (default: "notes")
- `MONGODB_COLLECTION` - The name of the MongoDB collection (default: "notes")

## 🧾 About TemplateTasks

TemplateTasks is a developer-focused initiative by Vadim Starichkov, currently operated as sole proprietorship in Finland.  
All code is released under open-source licenses. Ownership may be transferred to a registered business entity in the future.

## 📄 License & Attribution

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE.md) file for details.

### Using This Project?

If you use this code in your own projects, attribution is required under the MIT License:

```
Based on golang-simple-notes by Vadim Starichkov, TemplateTasks

https://github.com/starichkov/golang-simple-notes
```

**Copyright © 2025 Vadim Starichkov, TemplateTasks**
