package db

import "splitnow/auth-service/internal/types"
 
type Store struct{}

func (Store) CreateUser(email, username, password string) (types.User, error) {
	return CreateUser(email, username, password) 
} 

