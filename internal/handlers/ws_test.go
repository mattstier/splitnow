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

// test that the /ws endpoint upgrades an HTTP request to a websocket
func TestWSHandshake(t *testing.T) {
	store := mocks.NewStore(t)
	r := gin.New()
	r.GET("/ws", WS(store, nil)) // redis client is not used here

	server := httptest.NewServer(r)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if !assert.NoError(t, err) {
		return
	}
	defer conn.Close()

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	assert.Equal(t, "websocket", resp.Header.Get("Upgrade"))
}
