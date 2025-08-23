# Use the official Go image as a build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Copy go mod and download deps first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o chatserver ./cmd/server

# Final lightweight image
FROM debian:bookworm-slim

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/chatserver .

# Copy static files (client.html, etc.)
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./chatserver"]