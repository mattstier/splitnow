package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"splitnow/internal/handlers/mocks"
)

// starts a gin router with the ws handler and returns a connected websocket
func setupWS(t *testing.T, store Store) *websocket.Conn {
	t.Helper()
	r := gin.New()
	r.GET("/ws", WS(store, nil)) // redis client is not used here

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
	conn := setupWS(t, store)
	_ = conn
}

// test that subscribing to a nonexistent room returns an error frame
func TestWSSubscribeNonexistentRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(999).
		Return(false, nil)

	conn := setupWS(t, store)

	err := conn.WriteJSON(msgFrame{Type: "subscribe", Room: 999})
	assert.NoError(t, err)

	var got errorFrame
	err = conn.ReadJSON(&got)
	assert.NoError(t, err)

	assert.Equal(t, "error", got.Type)
	assert.Equal(t, 999, got.Room)
	//assert.Equal(t, "room not found", got.Message)
}
