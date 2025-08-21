# Gomegle 🎭

_A simple Omegle-style random chat server built with Go + WebSockets._

## ✨ Features

- Real-time chat using WebSockets
- Random pairing of anonymous users
- Skip/Next support
- Idle timeout to auto-disconnect inactive users
- Metrics endpoint for monitoring
- Redis Pub/Sub for scaling across multiple servers
- Docker & Docker Compose setup

## 🏗 Architecture

- Each server maintains a local Hub (WebSocket clients)
- Redis Pub/Sub syncs events (matchmaking, messages) across servers
- Stateless design — works behind load balancers
- Easy to scale horizontally

## 🚀 Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Redis (for cluster mode)
