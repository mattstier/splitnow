# SplitNow

Bill-splitting chatroom app.

# Architecture

To support (horizontal) scalability, modifiability, as well as fault-tolerance, a microservices architecture was chosen (this is not yet implemented). The backend is written in Go, using the Gin web framework and combines RESTful endpoints with WebSockets to deliver live messages, and load old ones on-the-go. Redis is used as a message broker to decouple message senders and receivers by enabling subscribing and publishing to 'rooms' - directly mapping to in-app chatrooms - and to enable one WebSocket connection to handle multiple message sources (chats). To handle message persistence PostgreSQL was chosen due to the relational nature of messages and the comprehensive Postgres ecosystem. Messages are saved by the backend every time before broadcasting them, serving as the ground-truth for messages, and ensuring that they are not accidentally lost if broadcasting fails.

## Project layout

```
.
├── db                   Postgres connection + queries
├── frontend             React + Vite app
├── internal/handlers    REST + WebSocket handlers
├── main.go              entrypoint, routes, config
└── sql                  schema migrations
```

# Development

## Dependencies

- Go 1.25+
- Node.js + npm
- PostgreSQL
- Redis (via Docker)

## Setup

```bash
# Database
sudo systemctl start postgresql

# Redis (Docker container)
docker start splitnow-redis   # create once with: docker run -d --name splitnow-redis -p 6379:6379 redis:8.10

# Backend
go mod tidy
go run .

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

Backend runs on `:8080`, frontend on `:5173`, Redis on `:6379`.

## Vite proxy

`vite.config.js` proxies `/ws` to the Go backend. The React code connects to `localhost:5173/ws` — Vite forwards the WebSocket to `localhost:8080/ws`. No CORS config needed.

## Testing API 
```bash
# install the test tooling (once)
go install github.com/vektra/mockery/v2@latest
go install gotest.tools/gotestsum@latest

# generate mocks from the Store interface
cd internal/handlers && mockery --name Store --dir ./

# run tests, colorized with gotestsum
cd internal/handlers && gotestsum --format testname

# or simply 
go test ./internal/handlers -v 

# or 
go test ./...
```

## Testing WebSocket directly

A single connection can subscribe to and leave multiple rooms via JSON frames:

```bash
npx wscat -c ws://localhost:8080/ws
> {"type":"subscribe","room":1}
> {"type":"send","room":1,"content":"hello"}
> {"type":"unsubscribe","room":1}
```

## Kill port

```bash
fuser -k 8080/tcp
```
