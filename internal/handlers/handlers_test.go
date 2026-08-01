package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"splitnow/internal/handlers/mocks"
	"splitnow/internal/types"
)

func TestCreateRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := mocks.NewStore(t)
	store.EXPECT().
		CreateRoom("gym", 1).
		Return(types.Room{ID: 1, Name: "gym", CreatedBy: 1}, nil)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"name":"gym","created_by":1}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var room types.Room
	err := json.Unmarshal(w.Body.Bytes(), &room)
	assert.NoError(t, err)
	assert.Equal(t, "gym", room.Name)
	assert.Equal(t, 1, room.CreatedBy)
}

func TestCreateRoomBadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
