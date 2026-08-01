package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// chatRoom is the set of clients subscribed to one room
type chatRoom struct {
	mu    sync.Mutex
	conns map[*wsClient]struct{}
}

// client is one websocket connection and the rooms it is subscribed to
type wsClient struct {
	mu    sync.Mutex
	conn  *websocket.Conn
	rooms map[int]struct{}
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
	room := &chatRoom{conns: make(map[*wsClient]struct{})}
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

func WS(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &wsClient{conn: conn, rooms: make(map[int]struct{})}
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
				// only allow subscribing to existing rooms
				exists, err := s.RoomExists(frame.Room)
				if err != nil {
					println("failed to check room:", err)
					continue
				}
				if !exists {
					data, _ := json.Marshal(errorFrame{Type: "error", Room: frame.Room, Message: "room not found"})
					conn.WriteMessage(websocket.TextMessage, data)
					continue
				}
				room := getChatRoom(frame.Room)
				client.mu.Lock()
				client.rooms[frame.Room] = struct{}{}
				client.mu.Unlock()
				room.mu.Lock()
				room.conns[client] = struct{}{}
				room.mu.Unlock()
				println("subscribed to room:", frame.Room)

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
