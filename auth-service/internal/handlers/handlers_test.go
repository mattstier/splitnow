package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"splitnow/auth-service/internal/handlers/mocks"
	"splitnow/auth-service/internal/types"
)

// test setup
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

//===== Tests for POST "/users" =====//

// testing user registration REST endpoint (positive case)
func TestCreateUser(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		CreateUser("example@gmail.com", "Foo", mock.Anything).
		Return(types.User{
			ID:        1,
			Email:     "example@gmail.com",
			Username:  "Foo",
			CreatedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		}, nil)

	router := gin.Default()
	router.POST("/users", CreateUser(store))

	body := bytes.NewBufferString(`{"email":"example@gmail.com","username":"Foo","password":"1234"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var user types.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)

	// check if returned fields are correct
	assert.Equal(t, "example@gmail.com", user.Email)
	assert.Equal(t, "Foo", user.Username)
	assert.Equal(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), user.CreatedAt)

	// check that it does not return the password in any way
	assert.NotContains(t, w.Body.String(), "password")
}
