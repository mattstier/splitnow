# SplitNow

Bill-splitting chatroom app.

# Dependencies

- Go 1.25+
- Node.js + npm

# Setup

```bash
# Backend
go mod tidy
go run .

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

Backend runs on `:8080`, frontend on `:5173`.

# Vite proxy

`vite.config.js` proxies `/ws` to the Go backend. The React code connects to `localhost:5173/ws` — Vite forwards the WebSocket to `localhost:8080/ws`. No CORS config needed.

# Testing WebSocket directly

```bash
npx wscat -c ws://localhost:8080/ws?room=test
```

# Kill port

```bash
fuser -k 8080/tcp
```
