package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"splitnow/internal/types"
)

type Store interface {
	CreateMessage(roomID int, sender, content string) (types.Message, error)
	GetMessagesByRoom(roomID int) ([]types.Message, error)
	CreateRoom(name string, creator int) (types.Room, error)
	GetAllRooms() ([]types.Room, error)
	GetRoomsWithName(name string) ([]types.Room, error)
	RoomExists(roomID int)(bool, error)
}

func Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// gets all messages by roomID
func GetMessages(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID, err := strconv.Atoi(c.Query("room"))
		if err != nil {
			c.JSON(400, gin.H{"error": "roomID required"})
			return
		}
		messages, err := s.GetMessagesByRoom(roomID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch messages"})
			return
		}
		c.JSON(200, messages)
	}
}

// create a room with a name and a creator with userID
func CreateRoom(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name      string `json:"name" binding:"required"`
			CreatedBy int    `json:"created_by" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid body"})
			return
		}

		room, err := s.CreateRoom(input.Name, input.CreatedBy)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to create room"})
			return
		}
		c.JSON(201, room)
	}
}

// gets all rooms (optionally filtered by name)
func GetRooms(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if roomName := c.Query("name"); roomName != "" {
			rooms, err := s.GetRoomsWithName(roomName)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to fetch rooms"})
				return
			}
			c.JSON(200, rooms)
			return
		}

		rooms, err := s.GetAllRooms()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch rooms"})
			return
		}
		c.JSON(200, rooms)
	}
}
