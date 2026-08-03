package handlers

import (
	"context"
	"encoding/json"
	"errors"
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
func TestWSSendMessage(t *testing.T) {
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

// test that sending without subscribing first returns an error frame
func TestWSSendMessageNotSubscribed(t *testing.T) {
	store := mocks.NewStore(t)
	conn := setupWS(t, store, nil) // path never touches redis

	err := conn.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})
	assert.NoError(t, err)

	var got errorFrame
	err = conn.ReadJSON(&got)
	assert.NoError(t, err)
	assert.Equal(t, "error", got.Type)
	assert.Equal(t, 1, got.Room)
	//assert.Equal(t, "not subscribed to room", got.Message)
}

// test that if a sent message could not be saved due to a store error, it does not crash
// TODO: notify the sender on save failure
func TestWSSendMessageStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		RoomExists(1).
		Return(true, nil) // needed for the subscribe step
	store.EXPECT().
		CreateMessage(1, "John Doe", "hello").
		Return(types.Message{}, errors.New("db down"))

	_, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)

	// listen on the room's redis channel from the test side
	sub := rdb.Subscribe(context.Background(), "room:1")
	t.Cleanup(func() { sub.Close() })
	ch := sub.Channel()

	// must be subscribed before sending
	err := conn.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	assert.NoError(t, err)

	err = conn.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		t.Fatalf("unexpected publish after failed save: %s", msg.Payload)
	case <-time.After(200 * time.Millisecond):
		// expected: store error is swallowed, nothing published
		// TODO: notify the sender on save failure
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

// test that a store error while checking the room is swallowed and no subscription opens
func TestWSSubscribeStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().RoomExists(1).Return(false, errors.New("db down"))

	mr, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)

	conn.WriteJSON(msgFrame{Type: "subscribe", Room: 1})

	assert.Eventually(t, func() bool {
		return mr.PubSubNumSub("room:1")["room:1"] == 0
	}, time.Second, 10*time.Millisecond)
}

// test that a frame that is not valid json is ignored and the connection stays alive
func TestWSInvalidFrame(t *testing.T) {
	store := mocks.NewStore(t)
	conn := setupWS(t, store, nil) // path never touches redis

	err := conn.WriteMessage(websocket.TextMessage, []byte("not json"))
	assert.NoError(t, err)

	// a valid frame after the bad one should still be processed
	err = conn.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})
	assert.NoError(t, err)

	var got errorFrame
	err = conn.ReadJSON(&got)
	assert.NoError(t, err)
	assert.Equal(t, "error", got.Type)
	assert.Equal(t, 1, got.Room)
}

// test that an unknown frame type is ignored and the connection stays alive
func TestWSUnknownFrameType(t *testing.T) {
	store := mocks.NewStore(t)
	conn := setupWS(t, store, nil) // path never touches redis

	err := conn.WriteJSON(msgFrame{Type: "bogus", Room: 1})
	assert.NoError(t, err)

	// a valid frame after the unknown one should still be processed
	err = conn.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})
	assert.NoError(t, err)

	var got errorFrame
	err = conn.ReadJSON(&got)
	assert.NoError(t, err)
	assert.Equal(t, "error", got.Type)
	assert.Equal(t, 1, got.Room)
}

func TestWSUnsubscribe(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().RoomExists(1).Return(true, nil) // needed for the subscribe step

	mr, rdb := newTestRedis(t)
	conn := setupWS(t, store, rdb)

	// get the room into a subscribed state first
	conn.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	assert.Eventually(t, func() bool {
		return mr.PubSubNumSub("room:1")["room:1"] == 1
	}, time.Second, 10*time.Millisecond)

	// unsubscribe: leaving the last member must drop the redis subscription
	conn.WriteJSON(msgFrame{Type: "unsubscribe", Room: 1})
	assert.Eventually(t, func() bool {
		return mr.PubSubNumSub("room:1")["room:1"] == 0
	}, time.Second, 10*time.Millisecond)
}

// test that a message is delivered to other room members, but not echoed to the sender
func TestWSDeliveryNoEcho(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().RoomExists(1).Return(true, nil).Times(2)
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
	sender := setupWS(t, store, rdb)
	receiver := setupWS(t, store, rdb)

	// both clients must subscribe before the message is sent
	err := sender.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	assert.NoError(t, err)
	err = receiver.WriteJSON(msgFrame{Type: "subscribe", Room: 1})
	assert.NoError(t, err)

	err = sender.WriteJSON(msgFrame{Type: "send", Room: 1, Content: "hello"})
	assert.NoError(t, err)

	// the other member receives the message
	var got types.Message
	receiver.SetReadDeadline(time.Now().Add(time.Second))
	err = receiver.ReadJSON(&got)
	assert.NoError(t, err)
	assert.Equal(t, "hello", got.Content)

	// the sender does not get an echo of its own message
	sender.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var echo types.Message
	err = sender.ReadJSON(&echo)
	assert.Error(t, err, "sender should not receive its own message")
}
