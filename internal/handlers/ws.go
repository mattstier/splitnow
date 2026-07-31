package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type chatRoom struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

// global map of all chatrooms and its mutex lock
var chatRooms = make(map[string]*chatRoom)
var chatRoomsMu sync.Mutex

// function creating a chatroom (returned by reference)
func getChatRoom(name string) *chatRoom {
	chatRoomsMu.Lock()
	defer chatRoomsMu.Unlock()
	if r, ok := chatRooms[name]; ok {
		return r
	}

	// chatroom initialized with an empty map of websocket.Conn
	r := &chatRoom{conns: make(map[*websocket.Conn]struct{})}
	chatRooms[name] = r
	return r
}

func WS(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID, err := strconv.Atoi(c.Query("room"))

		if err != nil {
			c.JSON(400, gin.H{"error": "roomID required"})
			return
		}

		// check if the room exists before connecting to it
		exists, err := s.RoomExists(roomID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to check room"})
			return
		}
		if !exists {
			c.JSON(404, gin.H{"error": "room not found"})
			return
		}

		// the upgrader is reponsible for switching this HTTP connection to a WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		room := getChatRoom(strconv.Itoa(roomID))

		// NOTE: you need to lock and unlock the room's mutex lock
		// so that the concurrent websockets don't hit race conditions on the connection list
		room.mu.Lock()
		room.conns[conn] = struct{}{}
		room.mu.Unlock()

		println("connected to room:", roomID)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				// if there is an error delete the connection i.e. disconnect
				room.mu.Lock()
				delete(room.conns, conn)
				room.mu.Unlock()
				println("disconnected from room:", roomID)
				break
			}

			// save message to db first in a blocking way
			saved, errm := s.CreateMessage(roomID, "John Doe", string(msg))
			if errm != nil {
				println("Failed to send message, reason: ", errm)
				continue
			}

			println("received:", string(msg))

			// broadcast the saved message as JSON to everyone (else) in the same room
			data, _ := json.Marshal(saved)
			room.mu.Lock()
			for other := range room.conns {
				if other != conn {
					other.WriteMessage(websocket.TextMessage, data)
				}
			}
			room.mu.Unlock()
		}
	}
}
