# Frontend

React + Vite app for Splitnow. State lives in **zustand** stores, one shared **axios** instance handles HTTP, and a WebSocket streams live messages.

## Architecture

The frontend is split into layers to separate concerns 


App.jsx      main application, top-level router 
Api.js       Axios instance + interceptors 
services/    Stateless API/WebSocket calls
stores/      Application state and actions that call services
views/       screen-level components

**Data flow**: a view -> store action -> service -> API -> API response -> store 

## Project layout

```
.
├── src/Api.js           Axios instance + interceptors
├── src/main.jsx         React entrypoint
├── src/App.jsx          main application, top-level router 
├── src/services/        endpoint wrappers with basic logic
├── src/stores/          Zustand stores, managing application state
├── src/views/           The different app "pages" (large React components) 
├── src/components/      Small, reusable React components (e.g.: styled buttons, input field)
├── src/utils/           standalone helpers 
└── vite.config.js       Vite dev proxy
```

## State management

`zustand` manages the state, allowing different views to use the same state without coupling it to one particular view (see: `useAuthStore(s => s.token)`). 

## Setup

```bash
cd frontend && npm install && npm run dev
Vite proxy
vite.config.js proxies the API so the browser only talks to the frontend origin, without CORS or hardcoded URLs. 
/login  /users        :8081 (auth service, ReST)
/rooms  /messages     :8080 (chat service, ReST)
/ws                   :8080 (chat service, WebSocket)
Environment vars
- VITE_API_ENDPOINT -> API base URL, default: relative to the same origin
```
