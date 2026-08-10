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

// creates a room with a given name, and creator
// and adds creator to the membership table in the same transaction
// if either fails, both are rolled back
func CreateRoom(name string, creator int) (types.Room, error) {
	// start transaction,
	tx, err := Pool.Begin(context.Background())
	if err != nil {
		return types.Room{}, err
	}
	// rollback if it fails before the commit
	defer tx.Rollback(context.Background())

	// create room in db, return it as Room struct
	row := tx.QueryRow(context.Background(),
		`INSERT INTO rooms (name, created_by)
         VALUES ($1, $2)
         RETURNING id, name, created_by, created_at`,
		name, creator)

	var r types.Room
	if err := row.Scan(&r.ID, &r.Name, &r.CreatedBy, &r.CreatedAt); err != nil {
		return types.Room{}, err
	}

	// creator added their own room (that the queryRow returned) as a member, idempotent
	_, err = tx.Exec(context.Background(),
		`INSERT INTO memberships (room_id, user_id)
         VALUES ($1, $2)
         ON CONFLICT DO NOTHING`,
		r.ID, creator)
	if err != nil {
		return types.Room{}, err
	}

	// applies transaction changes
	if err := tx.Commit(context.Background()); err != nil {
		return types.Room{}, err
	}
	return r, nil
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
