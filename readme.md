# Gomegle 🎭

_A simple Omegle-style random chat server built with Go + WebSockets._

## ✨ Features

- Real-time chat using WebSockets
- Random pairing of anonymous users
- Skip/Next support
- Idle timeout to auto-disconnect inactive users
- Metrics endpoint for monitoring

## 🏗 Architecture

- Each server maintains a local Hub (WebSocket clients)
- Stateless design — works behind load balancers
- Easy to scale horizontally

## To Build Image

docker build -t gomegle .

## To Run the Image

docker run -p 8080:8080 gomegle
