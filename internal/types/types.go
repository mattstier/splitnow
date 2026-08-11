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

var ErrRoomAlreadyJoined = errors.New("room already joined")
var ErrNotAMember = errors.New("the user is not a member of this room")
