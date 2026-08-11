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
	"splitnow/internal/token"
	"splitnow/internal/types"
)

// test setup
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// fakeAuth replaces RequireAuth in handler tests, setting the claims in the
// context just like the WS setup does
func fakeAuth(claims *token.Claims) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", claims)
		c.Next()
	}
}

//===== Tests for POST "/rooms" =====//

// testing room creation REST endpoint (positive case)
func TestCreateRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		CreateRoom("gym", 1).
		Return(types.Room{ID: 1, Name: "gym", CreatedBy: 1}, nil)

	router := gin.Default()
	router.POST("/rooms",
		fakeAuth(&token.Claims{UserID: 1, Username: "John Doe"}),
		CreateRoom(store))

	body := bytes.NewBufferString(`{"name":"gym"}`)
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

// testing room creation REST endpoint with empty name field (negative case)
func TestCreateRoomEmptyName(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/rooms", CreateRoom(store))

	body := bytes.NewBufferString(`{"name": ""}`)
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

// testing room creation REST endpoint when the store/database fails (negative case)
func TestCreateRoomStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	// introduce error in db
	store.EXPECT().
		CreateRoom("gym", 1).
		Return(types.Room{}, errors.New("database unavailable"))

	router := gin.Default()
	router.POST("/rooms",
		fakeAuth(&token.Claims{UserID: 1, Username: "John Doe"}),
		CreateRoom(store))

	body := bytes.NewBufferString(`{"name":"gym"}`)
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
		IsMember(42, 1).
		Return(true, nil)
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

	router.GET("/messages",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetMessages(store))

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
	store.EXPECT().
		IsMember(42, 999).
		Return(true, nil)
	// mock no match for a nonexistent room
	store.EXPECT().
		GetMessagesByRoom(999).
		Return([]types.Message{}, nil)

	router := gin.Default()
	router.GET("/messages",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetMessages(store))

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
	store.EXPECT().
		IsMember(42, 1).
		Return(true, nil)

	// introduce error in db
	store.EXPECT().
		GetMessagesByRoom(1).
		Return([]types.Message{}, errors.New("database unavailable"))

	router := gin.Default()
	router.GET("/messages",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetMessages(store))

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
	router.GET("/messages",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetMessages(store))

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

//===== Tests for POST "/rooms/:room_id/join" =====//

// testing AddMember REST endpoint (positive case)
func TestAddMember(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		AddMember(42, 1).
		Return(types.Membership{RoomID: 1, UserID: 42}, nil)

	router := gin.New()
	router.POST("/rooms/:room_id/join",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		AddMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var membership types.Membership
	err := json.Unmarshal(w.Body.Bytes(), &membership)
	assert.NoError(t, err)
	assert.Equal(t, types.Membership{RoomID: 1, UserID: 42}, membership)
}

// testing AddMember with an non-existent roomID (negative case)
func TestAddMemberRoomDoesNotExist(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(999).
		Return(false, nil)

	router := gin.New()
	router.POST("/rooms/:room_id/join",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		AddMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/999/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// testing AddMember with an invalid roomID (negative case)
func TestAddMemberInvalidRoomIDFormat(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.New()
	router.POST("/rooms/:room_id/join",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		AddMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/foo/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing AddMember idempotency
func TestAddMemberAlreadyJoined(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		AddMember(42, 1).
		Return(types.Membership{}, types.ErrRoomAlreadyJoined)

	router := gin.New()
	router.POST("/rooms/:room_id/join",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		AddMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// testing AddMember with a store error in the chat service
func TestAddMemberStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		AddMember(42, 1).
		Return(types.Membership{}, errors.New("database unavailable"))

	router := gin.New()
	router.POST("/rooms/:room_id/join",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		AddMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// testing AddMember with a store error when fetching rooms
func TestAddMemberRoomExistsError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(false, errors.New("database unavailable"))

	router := gin.New()
	router.POST("/rooms/:room_id/join",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		AddMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

//===== Tests for POST "/rooms/:room_id/leave" =====//

// testing RemoveMember (positive case)
func TestRemoveMember(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		RemoveMember(42, 1).
		Return(types.Membership{RoomID: 1, UserID: 42}, nil)

	router := gin.New()
	router.POST("/rooms/:room_id/leave",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		RemoveMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var membership types.Membership
	err := json.Unmarshal(w.Body.Bytes(), &membership)
	assert.NoError(t, err)
	assert.Equal(t, types.Membership{RoomID: 1, UserID: 42}, membership)
}

// testing RemoveMember with a non-existent roomID (negative case)
func TestRemoveMemberRoomDoesNotExist(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(999).
		Return(false, nil)

	router := gin.New()
	router.POST("/rooms/:room_id/leave",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		RemoveMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/999/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// testing RemoveMember with an invalid roomID (negative case)
func TestRemoveMemberInvalidRoomIDFormat(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.New()
	router.POST("/rooms/:room_id/leave",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		RemoveMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/foo/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// testing RemoveMember when the user is not a member of the room
func TestRemoveMemberNotAMember(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		RemoveMember(42, 1).
		Return(types.Membership{}, types.ErrNotAMember)

	router := gin.New()
	router.POST("/rooms/:room_id/leave",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		RemoveMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// testing RemoveMember with a store error in the chat service
func TestRemoveMemberStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		RemoveMember(42, 1).
		Return(types.Membership{}, errors.New("database unavailable"))

	router := gin.New()
	router.POST("/rooms/:room_id/leave",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		RemoveMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// testing RemoveMember with a store error when fetching rooms
func TestRemoveMemberRoomExistsError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(false, errors.New("database unavailable"))

	router := gin.New()
	router.POST("/rooms/:room_id/leave",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		RemoveMember(store))

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

//===== Tests for GET "/rooms/mine" =====//

// test for getting the rooms of a given user (positive case)
func TestGetUserRooms(t *testing.T) {
	store := mocks.NewStore(t)

	store.EXPECT().
		GetUserRooms(42).
		Return([]types.Room{
			{ID: 1, Name: "room1", CreatedBy: 42},
			{ID: 2, Name: "room2", CreatedBy: 42}}, nil)

	router := gin.New()
	router.GET("/rooms/mine",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetUserRooms(store))

	req := httptest.NewRequest(http.MethodGet, "/rooms/mine", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var rooms []types.Room
	err := json.Unmarshal(w.Body.Bytes(), &rooms)
	assert.NoError(t, err)
	assert.Len(t, rooms, 2)

	assert.Contains(t, rooms, types.Room{
		ID:        1,
		Name:      "room1",
		CreatedBy: 42,
	})
	assert.Contains(t, rooms, types.Room{
		ID:        2,
		Name:      "room2",
		CreatedBy: 42,
	})
}

// test for getting the rooms of a given user during a store/db error
func TestGetUserRoomsStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		GetUserRooms(42).
		Return([]types.Room{}, errors.New("database unavailable"))

	router := gin.New()
	router.GET("/rooms/mine",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetUserRooms(store))

	req := httptest.NewRequest(http.MethodGet, "/rooms/mine", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUserRoomsNoResult(t *testing.T) {
	store := mocks.NewStore(t)
	// mock no match for a given room 'nonexistent'
	store.EXPECT().
		GetUserRooms(42).
		Return([]types.Room{}, nil)
	router := gin.Default()
	router.GET("/rooms/mine",
		fakeAuth(&token.Claims{UserID: 42, Username: "John Doe"}),
		GetUserRooms(store))
	// query nonexistent room
	req := httptest.NewRequest(http.MethodGet, "/rooms/mine", nil)
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
