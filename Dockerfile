# Use the official Golang image as the base image
FROM golang:1.24.3-alpine3.21 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notes-api .

# Use a minimal alpine image for the final image
FROM alpine:3.21

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/notes-api .

# Expose the ports
EXPOSE 8080 8081

# Set environment variables
ENV COUCHDB_URL=http://couchdb:5984
ENV COUCHDB_DB=notes

# Run the application
CMD ["./notes-api"]
