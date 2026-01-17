# Testing Guide

This document provides instructions on how to run tests for the Notes API.

## 🧪 Testing

### Running Tests

- **All tests**: `go test ./...`
- **Unit tests only**: `go test -short ./...` (skips integration tests that require Docker).
- **Specific package**: `go test ./storage`

### Integration Tests

Integration tests for CouchDB and MongoDB use `testcontainers-go` and are automatically skipped in `-short` mode.

These tests require Docker to be running on your machine. The project uses a shared test state to speed up container startup across different packages.
