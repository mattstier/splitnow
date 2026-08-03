package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

//===== Tests for POST "/rooms" =====//

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

// testing room creation REST endpoint when the store/database fails (negative case)
func TestCreateRoomStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	// introduce error in db
	store.EXPECT().
		CreateRoom("gym", 1).
		Return(types.Room{}, errors.New("database unavailable"))

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"name":"gym","created_by":1}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// assert that it gives correct 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== Tests for GET "/rooms" =====//

// test for getting all rooms (positive case)
func TestGetRooms(t *testing.T) {
	store := mocks.NewStore(t)
	// add 'room1' and 'room2' to mock,
	store.EXPECT().
		GetAllRooms().
		Return([]types.Room{
			{ID: 1, Name: "room1", CreatedBy: 1},
			{ID: 2, Name: "room2", CreatedBy: 1}}, nil)

	router := gin.Default()
	router.GET("/rooms", GetRooms(store))

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var rooms []types.Room
	err := json.Unmarshal(w.Body.Bytes(), &rooms)
	assert.NoError(t, err)
	// see if 2 results
	assert.Len(t, rooms, 2)

	// see if the 2 results are correct regardless of order
	assert.Contains(t, rooms, types.Room{ID: 1, Name: "room1", CreatedBy: 1})
	assert.Contains(t, rooms, types.Room{ID: 2, Name: "room2", CreatedBy: 1})
}

// test for getting all rooms with a given name (positive case)
func TestGetRoomsWithName(t *testing.T) {
	store := mocks.NewStore(t)
	// add 'room1' and 'room1' with a different ID to mock,
	store.EXPECT().
		GetRoomsWithName("room1").
		Return([]types.Room{
			{ID: 2, Name: "room1", CreatedBy: 2},
			{ID: 1, Name: "room1", CreatedBy: 1}}, nil)

	router := gin.Default()
	router.GET("/rooms", GetRooms(store))

	// query room1 specifically
	req := httptest.NewRequest(http.MethodGet, "/rooms?name=room1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var rooms []types.Room
	err := json.Unmarshal(w.Body.Bytes(), &rooms)
	assert.NoError(t, err)
	// see if both rooms show up
	assert.Len(t, rooms, 2)

	// see if the correct room is returned
	assert.Contains(t, rooms, types.Room{ID: 1, Name: "room1", CreatedBy: 1})
	assert.Contains(t, rooms, types.Room{ID: 2, Name: "room1", CreatedBy: 2})
}

// test for querying a room with a given name that does not exist in store
func TestGetRoomNoResult(t *testing.T) {
	store := mocks.NewStore(t)
	// mock no match for a given room 'nonexistent'
	store.EXPECT().GetRoomsWithName("nonexistent").Return([]types.Room{}, nil)
	router := gin.Default()
	router.GET("/rooms", GetRooms(store))

	// query nonexistent room
	req := httptest.NewRequest(http.MethodGet, "/rooms?name=nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// by design it should not give 404, just 200 and an empty list
	assert.Equal(t, http.StatusOK, w.Code)

	var rooms []types.Room
	err := json.Unmarshal(w.Body.Bytes(), &rooms)
	assert.NoError(t, err)

	// see if the list is indeed empty
	assert.Len(t, rooms, 0)
	assert.Equal(t, rooms, []types.Room{})
}

// testing room querying REST endpoint when the store/database fails (negative case)
func TestGetRoomStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	// introduce error in db
	store.EXPECT().
		GetAllRooms().
		Return([]types.Room{}, errors.New("database unavailable"))

	router := gin.Default()
	router.GET("/rooms", GetRooms(store))

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// assert that it gives correct 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== Tests for GET "/messages" =====//

// test getting all the messages of an existing room
func TestGetMessagesByRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		GetMessagesByRoom(1).
		Return([]types.Message{
			{
				ID:        1,
				RoomID:    1,
				Sender:    "1",
				Content:   "foo",
				CreatedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:        2,
				RoomID:    1,
				Sender:    "2",
				Content:   "bar",
				CreatedAt: time.Date(2026, 6, 1, 13, 0, 0, 0, time.UTC),
			},
		}, nil)

	router := gin.Default()
	router.GET("/messages", GetMessages(store))

	req := httptest.NewRequest(http.MethodGet, "/messages?room=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var messages []types.Message
	err := json.Unmarshal(w.Body.Bytes(), &messages)
	assert.NoError(t, err)

	assert.Contains(t, messages, types.Message{
		ID:        1,
		RoomID:    1,
		Sender:    "1",
		Content:   "foo",
		CreatedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
	})

	assert.Contains(t, messages, types.Message{
		ID:        2,
		RoomID:    1,
		Sender:    "2",
		Content:   "bar",
		CreatedAt: time.Date(2026, 6, 1, 13, 0, 0, 0, time.UTC),
	})
}

// test getting all the messages of a room that does not exist
func TestGetMessagesByRoomNonexistentRoom(t *testing.T) {
	store := mocks.NewStore(t)
	// mock no match for a nonexistent room
	store.EXPECT().
		GetMessagesByRoom(999).
		Return([]types.Message{}, nil)

	router := gin.Default()
	router.GET("/messages", GetMessages(store))

	// query a nonexistent room
	req := httptest.NewRequest(http.MethodGet, "/messages?room=999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// by design it gives 200 and an empty list (no RoomExists check yet)
	assert.Equal(t, http.StatusOK, w.Code)

	var messages []types.Message
	err := json.Unmarshal(w.Body.Bytes(), &messages)
	assert.NoError(t, err)

	// see if the list is indeed empty
	assert.Len(t, messages, 0)
	assert.Equal(t, messages, []types.Message{})
}

// test getting all the messages of a room with a failed Store
func TestGetMessagesByRoomStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	// introduce error in db
	store.EXPECT().
		GetMessagesByRoom(1).
		Return([]types.Message{}, errors.New("database unavailable"))

	router := gin.Default()
	router.GET("/messages", GetMessages(store))

	req := httptest.NewRequest(http.MethodGet, "/messages?room=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// assert that it gives correct 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// test getting all the messages of a room without specifying a roomID (negative case)
func TestGetMessagesByRoomMissingRoomID(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.GET("/messages", GetMessages(store))

	req := httptest.NewRequest(http.MethodGet, "/messages?room=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// assert that it gives a 400 Bad Request 
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing the name-filtered room query when the store fails (negative case)
func TestGetRoomsWithNameStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		GetRoomsWithName("room1").
		Return([]types.Room{}, errors.New("database unavailable"))

	router := gin.Default()
	router.GET("/rooms", GetRooms(store))

	req := httptest.NewRequest(http.MethodGet, "/rooms?name=room1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
