# Setup and Running Guide

This document provides detailed instructions on how to set up, compile, and run the Notes API.

## 🏗 Setup & Running

### Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/starichkov/golang-simple-notes.git
cd golang-simple-notes
go mod tidy
```

### Compilation

To compile the application with a reduced binary size:

```bash
go build -ldflags="-s -w" -o notes-api
```
- `-s`: removes the symbol table and debug info.
- `-w`: removes DWARF debugging information.

### Running with Docker Compose

The application provides several Docker Compose files for different storage backends:

#### In-Memory Storage
```bash
docker-compose up -d
```
Accessible at: REST `http://localhost:8080`, gRPC `localhost:8081`

#### CouchDB Storage
```bash
docker-compose -f docker-compose.couchdb.yml up -d
```
Accessible at: REST `http://localhost:8080`, gRPC `localhost:8081`, CouchDB `http://localhost:5984` (admin:password)

#### MongoDB Storage
```bash
docker-compose -f docker-compose.mongodb.yml up -d
```
Accessible at: REST `http://localhost:8080`, gRPC `localhost:8081`, MongoDB `mongodb://localhost:27017` (admin:password)

### Running Locally

You can also run the application directly using Go:

```bash
# In-memory (default)
go run .

# CouchDB
export STORAGE_TYPE=couchdb
export COUCHDB_URL=http://admin:password@localhost:5984
go run .

# MongoDB
export STORAGE_TYPE=mongodb
export MONGODB_URI=mongodb://admin:password@localhost:27017
go run .
```

## ⚙️ Configuration

The application is configured via environment variables:

| Variable             | Description                                        | Default                     |
|----------------------|----------------------------------------------------|-----------------------------|
| `STORAGE_TYPE`       | Type of storage: `memory`, `couchdb`, or `mongodb` | `memory`                    |
| `COUCHDB_URL`        | URL of the CouchDB server                          | `http://localhost:5984`     |
| `COUCHDB_DB`         | Name of the CouchDB database                       | `notes`                     |
| `MONGODB_URI`        | URI of the MongoDB server                          | `mongodb://localhost:27017` |
| `MONGODB_DB`         | Name of the MongoDB database                       | `notes`                     |
| `MONGODB_COLLECTION` | Name of the MongoDB collection                     | `notes`                     |

*Note: Ports are currently hardcoded to `:8080` (REST) and `:8081` (gRPC).*
