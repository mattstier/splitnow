package db

import (
	"context"
	"errors"
	"splitnow/auth-service/internal/types"

	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUser(email, username, password string) (types.User, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO users (email, username, password)
         VALUES ($1, $2, $3)
         RETURNING id, email, username, created_at`,
		email, username, password)

	var user types.User
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		// handle taken unique attributes errors
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return types.User{}, types.ErrEmailTaken
			case "users_username_key":
				return types.User{}, types.ErrUsernameTaken
			}
		}
		return types.User{}, err
	}
	return user, nil
}
