package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// chatRoom is the set of clients subscribed to one room
// It also owns the redis subscription for that room
type chatRoom struct {
	mu     sync.Mutex
	conns  map[*wsClient]struct{}
	id     int
	pubsub *redis.PubSub // pubsub is non-nil as long as at least one client is subscribed to the room
}

// client is one websocket connection and the rooms it is subscribed to
type wsClient struct {
	mu    sync.Mutex
	conn  *websocket.Conn
	rooms map[int]struct{}
}

type wsDeps struct {
	store       Store
	redisClient *redis.Client
}

// global map of all chatrooms and its mutex lock
var chatRooms = make(map[int]*chatRoom)
var chatRoomsMu sync.Mutex

// returns the chatroom for roomID, creating it if needed
func getChatRoom(roomID int) *chatRoom {
	chatRoomsMu.Lock()
	defer chatRoomsMu.Unlock()
	if room, ok := chatRooms[roomID]; ok {
		return room
	}
	room := &chatRoom{id: roomID, conns: make(map[*wsClient]struct{})}
	chatRooms[roomID] = room
	return room
}

// incoming frames from the client
type msgFrame struct {
	Type    string `json:"type"`
	Room    int    `json:"room"`
	Content string `json:"content"`
}

// outgoing error frame
type errorFrame struct {
	Type    string `json:"type"`
	Room    int    `json:"room"`
	Message string `json:"message"`
}

func (ws wsDeps) handleSubscribe(client *wsClient, frame msgFrame) {

	// only allow subscribing to existing rooms
	exists, err := ws.store.RoomExists(frame.Room)
	if err != nil {
		println("failed to check room:", err)
		return
	}
	if !exists {
		data, _ := json.Marshal(errorFrame{Type: "error", Room: frame.Room, Message: "room not found"})
		client.conn.WriteMessage(websocket.TextMessage, data)
		return
	}
	room := getChatRoom(frame.Room)
	room.mu.Lock()

	// first local member of this room opens the redis subscription so we
	// receive messages published by other instances too
	if room.pubsub == nil {
		room.pubsub = ws.redisClient.Subscribe(context.Background(), "room:"+strconv.Itoa(room.id))
		go room.consume()
	}
	room.conns[client] = struct{}{}
	room.mu.Unlock()
	client.mu.Lock()
	client.rooms[frame.Room] = struct{}{}
	client.mu.Unlock()
	println("subscribed to room:", frame.Room)
}

// consume runs in its own goroutine per room: every message published to the
// room's redis channel is forwarded to all locally connected clients.
// stops when the subscription is closed
func (room *chatRoom) consume() {
	ch := room.pubsub.Channel()
	for msg := range ch {
		room.mu.Lock()
		for other := range room.conns {
			other.conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		}
		room.mu.Unlock()
	}
	room.pubsub.Close()
}

func WS(s Store, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &wsClient{conn: conn, rooms: make(map[int]struct{})}
		deps := wsDeps{store: s, redisClient: redisClient}
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				// remove the client from every room it subscribed to
				client.mu.Lock()
				rooms := make([]int, 0, len(client.rooms))
				for roomID := range client.rooms {
					rooms = append(rooms, roomID)
				}
				client.mu.Unlock()
				for _, roomID := range rooms {
					room := getChatRoom(roomID)
					room.mu.Lock()
					delete(room.conns, client)
					room.mu.Unlock()
				}
				println("disconnected")
				break
			}

			var frame msgFrame
			if err := json.Unmarshal(msg, &frame); err != nil {
				println("invalid frame")
				continue
			}

			switch frame.Type {
			case "subscribe":
				deps.handleSubscribe(client, frame)
			case "unsubscribe":
				client.mu.Lock()
				delete(client.rooms, frame.Room)
				client.mu.Unlock()
				room := getChatRoom(frame.Room)
				room.mu.Lock()
				delete(room.conns, client)
				room.mu.Unlock()
				println("unsubscribed from room:", frame.Room)

			case "send":
				// must be subscribed to the room to send in it
				client.mu.Lock()
				_, ok := client.rooms[frame.Room]
				client.mu.Unlock()
				if !ok {
					data, _ := json.Marshal(errorFrame{Type: "error", Room: frame.Room, Message: "not subscribed to room"})
					conn.WriteMessage(websocket.TextMessage, data)
					continue
				}
				saved, errm := s.CreateMessage(frame.Room, "John Doe", frame.Content)
				if errm != nil {
					println("failed to save message:", errm)
					continue
				}
				data, _ := json.Marshal(saved)
				room := getChatRoom(frame.Room)
				room.mu.Lock()
				for other := range room.conns {
					if other != client {
						other.conn.WriteMessage(websocket.TextMessage, data)
					}
				}
				room.mu.Unlock()

			default:
				println("unknown frame type:", frame.Type)
			}
		}
	}
}
