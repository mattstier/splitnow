package db

import (
	"context"

	"github.com/jackc/pgx/v5"

	"splitnow/internal/types"
)

func CreateMessage(roomID int, sender, content string) (types.Message, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO messages (room_id, sender, content)
         VALUES ($1, $2, $3)
         RETURNING id, room_id, sender, content, deleted, created_at`,
		roomID, sender, content)

	var m types.Message
	err := row.Scan(&m.ID, &m.RoomID, &m.Sender, &m.Content, &m.Deleted, &m.CreatedAt)
	return m, err
}

func GetMessagesByRoom(roomID int) ([]types.Message, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, sender, content, deleted, created_at
         FROM messages WHERE room_id = $1 ORDER BY created_at`, roomID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[types.Message])
}
