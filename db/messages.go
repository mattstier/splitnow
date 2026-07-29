package db

import (
    "time"
)

type Message struct {
    ID        int
    RoomID    string
    Sender    string
    Content   string
    CreatedAt time.Time
}
