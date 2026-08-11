package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"splitnow/internal/token"
	"splitnow/internal/types"
)

type Store interface {
	CreateMessage(roomID int, sender, content string) (types.Message, error)
	GetMessagesByRoom(roomID int) ([]types.Message, error)
	CreateRoom(name string, creator int) (types.Room, error)
	GetAllRooms() ([]types.Room, error)
	GetRoomsWithName(name string) ([]types.Room, error)
	RoomExists(roomID int) (bool, error)
	AddMember(userID, roomID int) (types.Membership, error)
	IsMember(userID, roomID int) (bool, error)
	GetUserRooms(userID int) ([]types.Room, error)
	RemoveMember(userID, roomID int) (types.Membership, error)
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

		// get UID from the JWT, check if its a member
		claims := c.MustGet("user").(*token.Claims)
		isMember, err := s.IsMember(claims.UserID, roomID)

		if !isMember {
			// reject non-members before fetching any messages
			c.JSON(http.StatusForbidden, gin.H{"error": "failed to fetch messages"})
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
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid body"})
			return
		}

		claims := c.MustGet("user").(*token.Claims)
		room, err := s.CreateRoom(input.Name, claims.UserID)
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

// add a member to a room
func AddMember(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// getting the userID from the JWT directly
		claims := c.MustGet("user").(*token.Claims)
		roomID, err := strconv.Atoi(c.Param("room_id"))

		// check that param room is a valid int
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomID format"})
			return
		}

		roomExists, err := s.RoomExists(roomID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch room"})
			return
		}

		if !roomExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}

		membership, err := s.AddMember(claims.UserID, roomID)
		if err != nil {
			if errors.Is(err, types.ErrRoomAlreadyJoined) {
				c.JSON(http.StatusConflict, gin.H{"error": "room already joined"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
			return
		}
		c.JSON(http.StatusCreated, membership)
	}
}

func RemoveMember(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// getting the userID from the JWT directly
		claims := c.MustGet("user").(*token.Claims)
		roomID, err := strconv.Atoi(c.Param("room_id"))

		// check that param room is a valid int
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomID format"})
			return
		}

		roomExists, err := s.RoomExists(roomID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch room"})
			return
		}

		if !roomExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}

		membership, err := s.RemoveMember(claims.UserID, roomID)
		if err != nil {
			if errors.Is(err, types.ErrNotAMember) {
				c.JSON(http.StatusNotFound, gin.H{"error": "cannot remove non-member from a room"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
			return
		}
		c.JSON(http.StatusOK, membership)
	}
}

func GetUserRooms(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// getting the userID from the JWT directly
		claims := c.MustGet("user").(*token.Claims)
		rooms, err := s.GetUserRooms(claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rooms"})
			return
		}

		c.JSON(http.StatusOK, rooms)
	}
}
