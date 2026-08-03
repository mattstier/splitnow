package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"splitnow/internal/handlers/mocks"
	"splitnow/internal/types"
)

// newTestRedis starts an in-memory redis server and returns a new redis client
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// starts a gin router with the ws handler and returns a connected websocket
func setupWS(t *testing.T, store Store, rdb *redis.Client) *websocket.Conn {
	t.Helper()
	r := gin.New()
	r.GET("/ws", WS(store, rdb))

	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	assert.Equal(t, "websocket", resp.Header.Get("Upgrade"))
	t.Cleanup(func() { conn.Close() })
	return conn
}

// test that the /ws endpoint upgrades an HTTP request to a websocket
func TestWSHandshake(t *testing.T) {
	store := mocks.NewStore(t)
	conn := setupWS(t, store, newTestRedis(t))
	_ = conn
}

// test that subscribing to a valid room and then sending delivers the message
func TestWSSubscribeValidRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		CreateMessage(1, "John Doe", "hello").
		Return(types.Message{
			ID: 1, 
			RoomID: 1, 
			Sender: "John Doe", 
			Content: "hello", 
			CreatedAt: time.Now(),
		}, nil)

	rdb := newTestRedis(t)

	sender := setupWS(t, store, rdb)
	receiver := setupWS(t, store, rdb)

	// each client must subscribe before it can receive
	sender.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	receiver.WriteJSON(msgFrame{Type: "subscribe", Room: 1})

	// sender sends, receiver should receive it
	sender.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})

	var got types.Message
	err := receiver.ReadJSON(&got)
	assert.NoError(t, err)
	assert.Equal(t, "hello", got.Content)
}

// test that subscribing to a nonexistent room returns an error frame
func TestWSSubscribeNonexistentRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(999).
		Return(false, nil)

	conn := setupWS(t, store, newTestRedis(t))

	err := conn.WriteJSON(msgFrame{Type: "subscribe", Room: 999})
	assert.NoError(t, err)

	var got errorFrame
	err = conn.ReadJSON(&got)
	assert.NoError(t, err)

	assert.Equal(t, "error", got.Type)
	assert.Equal(t, 999, got.Room)
	//assert.Equal(t, "room not found", got.Message)
}
