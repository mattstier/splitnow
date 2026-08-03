package handlers

import (
	"context"
	"encoding/json"
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

// newTestRedis starts an in-memory redis server and returns the server plus a
// new redis client pointed at it
func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	return mr, redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// resetChatRooms clears the package-level room map between tests, so one test
// never sees a chatroom (or its redis subscription) left behind by another
func resetChatRooms() {
	chatRoomsMu.Lock()
	defer chatRoomsMu.Unlock()
	for id, room := range chatRooms {
		room.mu.Lock()
		if room.pubsub != nil {
			room.pubsub.Close()
		}
		room.mu.Unlock()
		delete(chatRooms, id)
	}
	connCounter.Store(0)
}

// starts a gin router with the ws handler and returns a connected websocket
func setupWS(t *testing.T, store Store, rdb *redis.Client) *websocket.Conn {
	t.Helper()
	t.Cleanup(resetChatRooms)

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
	_, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)
	_ = conn
}

// test that subscribing to a valid room opens a redis subscription for it
func TestWSSubscribeValidRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)

	mr, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)

	err := conn.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	assert.NoError(t, err)

	// subscribing produces no frame back, so assert on redis state: the
	// handler must have opened a subscription for room:1
	assert.Eventually(t, func() bool {
		return mr.PubSubNumSub("room:1")["room:1"] == 1
	}, time.Second, 10*time.Millisecond)
}

// test that sending a message publishes it to the room's redis channel
func TestWSSend(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil)
	store.EXPECT().
		CreateMessage(1, "John Doe", "hello").
		Return(types.Message{
			ID:        1,
			RoomID:    1,
			Sender:    "John Doe",
			Content:   "hello",
			CreatedAt: time.Now(),
		}, nil)

	_, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)

	// must be subscribed before sending
	err := conn.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	assert.NoError(t, err)

	// listen on the room's redis channel from the test side
	sub := rdb.Subscribe(context.Background(), "room:1")
	t.Cleanup(func() { sub.Close() })
	ch := sub.Channel()

	err = conn.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		var p pubPayload
		assert.NoError(t, json.Unmarshal([]byte(msg.Payload), &p))
		assert.Equal(t, "hello", p.Message.Content)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for published message")
	}
}

// test that subscribing to a nonexistent room returns an error frame
func TestWSSubscribeNonexistentRoom(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(999).
		Return(false, nil)

	_, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)

	err := conn.WriteJSON(msgFrame{Type: "subscribe", Room: 999})
	assert.NoError(t, err)

	var got errorFrame
	err = conn.ReadJSON(&got)
	assert.NoError(t, err)

	assert.Equal(t, "error", got.Type)
	assert.Equal(t, 999, got.Room)
	//assert.Equal(t, "room not found", got.Message)
}
