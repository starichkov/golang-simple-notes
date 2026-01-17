# Notes API: Golang with NoSQL Databases

[![Author](https://img.shields.io/badge/Author-Vadim%20Starichkov-blue?style=for-the-badge)](https://github.com/starichkov)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/starichkov/golang-simple-notes?style=for-the-badge)
[![GitHub License](https://img.shields.io/github/license/starichkov/golang-simple-notes?style=for-the-badge)](https://github.com/starichkov/golang-simple-notes/blob/main/LICENSE.md)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/starichkov/golang-simple-notes/build.yml?style=for-the-badge)](https://github.com/starichkov/golang-simple-notes/actions/workflows/build.yml)
[![Codecov](https://img.shields.io/codecov/c/github/starichkov/golang-simple-notes?style=for-the-badge)](https://codecov.io/gh/starichkov/golang-simple-notes)

A simple microservice for notes management with REST and gRPC APIs, supporting multiple storage backends.

*This project is generated using JetBrains Junie and several other AI coding agents to evaluate agent capabilities.*

## 🚀 Features

- **Dual API Support**: Full CRUD operations via both REST and gRPC.
- **Multiple Storage Backends**:
    - **In-memory**: Ideal for local development and testing.
    - **CouchDB**: Support for document-oriented storage with CouchDB.
    - **MongoDB**: Support for document-oriented storage with MongoDB.
- **Clean Architecture**: Decoupled domain logic, storage interfaces, and transport layers.
- **Dockerized**: Easy deployment with Docker and Docker Compose.
- **Comprehensive Testing**: Unit tests and integration tests using `testcontainers-go`.

## 🛠 Prerequisites

- **Go**: 1.25.6 or later (as specified in `go.mod`).
- **Docker & Docker Compose**: Required for running databases and integration tests.

## 📁 Project Structure

```text
.
├── grpc/           # gRPC service implementation
├── model/          # Domain entities (Note)
├── proto/          # gRPC service definitions (Protocol Buffers)
├── rest/           # REST API handlers and middleware
├── storage/        # Storage interface and implementations (Memory, CouchDB, MongoDB)
├── app.go          # Application wiring and lifecycle management
├── config.go       # Configuration management via environment variables
├── Dockerfile      # Docker image definition
├── docker-compose* # Docker Compose configurations for various setups
└── main.go         # Application entry point
```

## 📚 Documentation

For more detailed information, please refer to the following guides:

- [**Setup & Running**](RUNNING.md) - How to install, compile, and run the application.
- [**API Reference**](API.md) - Details about REST and gRPC endpoints.
- [**Testing**](TESTING.md) - How to run unit and integration tests.

## 🧾 About TemplateTasks

TemplateTasks is a developer-focused initiative by Vadim Starichkov, currently operated as sole proprietorship in Finland.  
All code is released under open-source licenses. Ownership may be transferred to a registered business entity in the future.

## 📄 License & Attribution

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE.md) file for details.

### Using This Project?

If you use this code in your own projects, attribution is required under the MIT License:

```
Based on golang-simple-notes by Vadim Starichkov, TemplateTasks

https://github.com/starichkov/golang-simple-notes
```

**Copyright © 2025 Vadim Starichkov, TemplateTasks**
