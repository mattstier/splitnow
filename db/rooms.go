package db

import (
	"context"
	"time"
)

type Room struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateRoom(name string, creator int) (Room, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO rooms (name, created_by)
         VALUES ($1, $2)
         RETURNING id, name, created_by, created_at`,
		name, creator)

	var r Room
	err := row.Scan(&r.ID, &r.Name, &r.CreatedBy, &r.CreatedAt)
	return r, err
}
