package handlers

import (
	"errors"
	"net/http"
	"splitnow/auth-service/internal/types"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	CreateUser(email, username, password string) (types.User, error)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// registers a new user and hashes the password using
func CreateUser(s Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required"`
			Username string `json:"username" binding:"required,min=3,max=20"`
			Password string `json:"password" binding:"required,min=8,max=72"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "invalid body"})
			return
		}

		hashedPassword, err := HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "failed to register user"})
			return
		}

		user, err := s.CreateUser(input.Email, input.Username, hashedPassword)
		if err != nil {
			// check error type and return the fitting message and status code
			switch {
			case errors.Is(err, types.ErrEmailTaken):
				c.JSON(http.StatusConflict,
					gin.H{"error": "email already taken", "field": "email"})
			case errors.Is(err, types.ErrUsernameTaken):
				c.JSON(http.StatusConflict,
					gin.H{"error": "username already taken", "field": "username"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
			}
			return
		}
		c.JSON(http.StatusCreated, user)
	}
}
