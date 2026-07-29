package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Message struct {
    ID        int
    RoomID    string
    Sender    string
    Content   string
    CreatedAt time.Time
}

func CreateMessage(roomID, sender, content string) (Message, error) {
    row := Pool.QueryRow(context.Background(),
        `INSERT INTO messages (room_id, sender, content)
         VALUES ($1, $2, $3)
         RETURNING id, room_id, sender, content, created_at`,
        roomID, sender, content)

    var m Message
    err := row.Scan(&m.ID, &m.RoomID, &m.Sender, &m.Content, &m.CreatedAt)
    return m, err
}

func GetMessagesByRoom(roomID string) ([]Message, error) {
    rows, err := Pool.Query(context.Background(),
        `SELECT id, room_id, sender, content, created_at
         FROM messages WHERE room_id = $1 ORDER BY created_at`, roomID)
    if err != nil {
        return nil, err
    }
    return pgx.CollectRows(rows, pgx.RowToStructByName[Message])
}
