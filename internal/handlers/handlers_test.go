package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"splitnow/internal/handlers/mocks"
	"splitnow/internal/types"
)

// test setup
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// testing room creation REST endpoint (positive case)
func TestCreateRoom(t *testing.T) {
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

// testing room creation REST endpoint with missing name field (negative case)
func TestCreateRoomMissingNameField(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"created_by": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing room creation REST endpoint with empty name field (negative case)
func TestCreateRoomEmptyName(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"created_by": 1, "name": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing room creation REST endpoint with an empty JSON (negative case)
func TestCreateRoomEmptyJSON(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing room creation REST endpoint with an invalid JSON format (negative case)
func TestCreateRoomNotJSON(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`"name": "foo"`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing room creation REST endpoint with missing creator field (negative case)
func TestCreateRoomMissingCreatorField(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"name": "foo"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TODO: revisit when creator existence is validated (auth service)
func TestCreateRoomNonexistentCreator(t *testing.T) {
	t.Skip("creator existence not validated yet, no users table until auth service")
}
