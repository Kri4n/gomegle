# Build stage
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /chat ./cmd/server


# Minimal runtime
FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web
EXPOSE 8080
ENTRYPOINT ["/chat"]