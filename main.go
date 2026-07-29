package main

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"splitnow/db"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Room struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

// global map of all chatrooms and its mutex lock
var rooms = make(map[string]*Room)
var roomsMu sync.Mutex

// function creating a chatroom (returned by reference)
func getRoom(name string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	if r, ok := rooms[name]; ok {
		return r
	}

	// chatroom initialized with an empty map of websocket.Conn
	r := &Room{conns: make(map[*websocket.Conn]struct{})}
	rooms[name] = r
	return r
}


func main() {
	// connnecting to the splitnow db with root user (for now)
	err := db.Connect("postgres://matestier@/splitnow?host=/var/run/postgresql")
	if err != nil {
		println("Failed to connect to db, reason: ", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})


	// gets all messages
	r.GET("/messages", func(c *gin.Context) {
		roomID := c.Query("room")
		if roomID == "" {
			c.JSON(400, gin.H{"error": "roomID required"})
			return
		}
		messages, err := db.GetMessagesByRoom(roomID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch messages"})
			return
		}
		c.JSON(200, messages)
	})

	r.GET("/ws", func(c *gin.Context) {
		roomID := c.Query("room")

		// the upgrader is reponsible for switching this HTTP connection to a WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		room := getRoom(roomID)

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
			_, errm := db.CreateMessage("test", "John Doe", string(msg))
			if errm != nil {
				println("Failed to send message, reason: ", errm)
			}

			println("received:", string(msg))

			room.mu.Lock()
			// broadcasts the message to everyone (else) in the same room
			for other := range room.conns {
				if other != conn {
					other.WriteMessage(websocket.TextMessage, msg)
				}
			}
			room.mu.Unlock()
		}
	})
	r.Run()
}
