package db

import (
	"context"
	"splitnow/auth-service/internal/types"
)

func CreateUser(email, username, password string) (types.User, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO users (email, username, password)
         VALUES ($1, $2, $3)
         RETURNING id, email, username, created_at`,
		email, username, password)

	var user types.User
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt)
	return user, err
}
