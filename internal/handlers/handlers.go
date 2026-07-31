package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"splitnow/db"
)

func Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// gets all messages by roomID
func GetMessages(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Query("room"))
	if err != nil {
		c.JSON(400, gin.H{"error": "roomID required"})
		return
	}
	messages, err := db.GetMessagesByRoom(roomID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch messages"})
		return
	}
	c.JSON(200, messages)
}

// create a room with a name and a creator with userID
func CreateRoom(c *gin.Context) {
	var input struct {
		Name      string `json:"name"`
		CreatedBy int    `json:"created_by"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	room, err := db.CreateRoom(input.Name, input.CreatedBy)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create room"})
		return
	}
	c.JSON(201, room)
}

// gets all rooms (optionally filtered by name)
func GetRooms(c *gin.Context) {
	if roomName := c.Query("name"); roomName != "" {
		rooms, err := db.GetRoomsWithName(roomName)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch rooms"})
			return
		}
		c.JSON(200, rooms)
		return
	}

	rooms, err := db.GetAllRooms()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch rooms"})
		return
	}
	c.JSON(200, rooms)
}
