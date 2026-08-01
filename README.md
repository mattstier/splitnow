# SplitNow

Bill-splitting chatroom app.

# Dependencies

- Go 1.25+
- Node.js + npm
- PostgreSQL
- Redis (via Docker)

# Setup

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

# Vite proxy

`vite.config.js` proxies `/ws` to the Go backend. The React code connects to `localhost:5173/ws` — Vite forwards the WebSocket to `localhost:8080/ws`. No CORS config needed.

# Testing WebSocket directly

A single connection can subscribe to and leave multiple rooms via JSON frames:

```bash
npx wscat -c ws://localhost:8080/ws
> {"type":"subscribe","room":1}
> {"type":"send","room":1,"content":"hello"}
> {"type":"unsubscribe","room":1}
```

# Kill port

```bash
fuser -k 8080/tcp
```
