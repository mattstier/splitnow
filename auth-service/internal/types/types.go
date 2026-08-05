package types

import (
	"errors"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrEmailTaken = errors.New("email already taken")
var ErrUsernameTaken = errors.New("username already taken")
