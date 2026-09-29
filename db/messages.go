package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"splitnow/internal/types"
)

func CreateMessage(roomID int, sender, content string, replyTo *int) (types.Message, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO messages (room_id, sender, content, reply_to)
         VALUES ($1, $2, $3, $4)
         RETURNING id, room_id, sender, content, deleted, created_at, reply_to`,
		roomID, sender, content, replyTo)

	var m types.Message
	err := row.Scan(&m.ID, &m.RoomID, &m.Sender, &m.Content, &m.Deleted, &m.CreatedAt, &m.ReplyTo)
	return m, err
}

func GetMessagesByRoom(roomID, messageID, limit int) ([]types.Message, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, sender, content, deleted, created_at, reply_to
         FROM messages 
		 WHERE room_id = $1 AND ($2 = 0 OR id < $2)
		 ORDER BY id DESC
		 LIMIT $3`, roomID, messageID, limit)

	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[types.Message])
}

// messages are deleted primarily by their own id, however roomID is required
// for authorization and sender to ensure only the owner can delete the message
func DeleteMessage(messageID, roomID int, sender string) (types.Message, error) {
	//NOTE: messages are not physically deleted, only marked as deleted and scrubbed of their content
	row := Pool.QueryRow(context.Background(),
		`UPDATE messages
		 SET deleted = TRUE, content = ''
		 WHERE id = $1 and room_id = $2 and sender = $3
         RETURNING id, room_id, sender, content, deleted, created_at, reply_to`,
		messageID, roomID, sender)

	var m types.Message
	err := row.Scan(&m.ID, &m.RoomID, &m.Sender, &m.Content, &m.Deleted, &m.CreatedAt, &m.ReplyTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Message{}, types.ErrMessageNotFound
	}
	return m, err
}
