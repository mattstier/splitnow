package types

import (
	"errors"
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"created_at"`
}

type Room struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Membership struct {
	UserID int `json:"user_id"`
	RoomID int `json:"room_id"`
}

// HATEOAS specific structs
type MessagePageLinks struct {
	Self string `json:"self"`
	Next string `json:"next,omitempty"`
}

type MessagePage struct {
	Data  []Message        `json:"data"`
	Links MessagePageLinks `json:"links"`
}

var ErrRoomAlreadyJoined = errors.New("room already joined")
var ErrNotAMember = errors.New("the user is not a member of this room")
var ErrMessageNotFound = errors.New("message not found")
