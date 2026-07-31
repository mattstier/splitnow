package db

import (
	"context"

	"github.com/jackc/pgx/v5"

	"splitnow/internal/types"
)

// check if a room with id exists
func RoomExists(roomID int) (bool, error) {
	var exists bool
	err := Pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`, roomID).Scan(&exists)
	return exists, err
}

func CreateRoom(name string, creator int) (types.Room, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO rooms (name, created_by)
         VALUES ($1, $2)
         RETURNING id, name, created_by, created_at`,
		name, creator)

	var r types.Room
	err := row.Scan(&r.ID, &r.Name, &r.CreatedBy, &r.CreatedAt)
	return r, err
}

func GetAllRooms() ([]types.Room, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT id, name, created_by, created_at
         FROM rooms`)

	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[types.Room])
}

func GetRoomsWithName(name string) ([]types.Room, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT id, name, created_by, created_at
         FROM rooms WHERE name = $1 ORDER BY created_at`, name)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[types.Room])
}
