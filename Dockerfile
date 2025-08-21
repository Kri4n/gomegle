FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-server ./cmd/server
FROM alpine:3.20
COPY --from=builder /app/server .
EXPOSE 5040
ENV PORT=5040
ENTRYPOINT ["./server"]