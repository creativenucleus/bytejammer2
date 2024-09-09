package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{}

func TestReconnect(t *testing.T) {
	require := require.New(t)

	wsTestServer, err := NewWsTestServer()
	require.NoError(err)
	defer wsTestServer.Close()

	//	wsTester.Start()

	conn, err := NewWebSocketConnection(wsTestServer.url)
	require.NoError(err)

	err = conn.SendRaw([]byte("hello"))
	require.NoError(err)

	wsTestServer.Close()

	time.Sleep(11 * time.Second)

	err = conn.SendRaw([]byte("hello2"))
	require.NoError(err)
	require.Error(err)
}

type wsTestServer struct {
	server *httptest.Server
	url    url.URL
}

func NewWsTestServer() (*wsTestServer, error) {
	wsts := wsTestServer{}

	// Create test server with the echo handler.
	wsts.server = httptest.NewServer(http.HandlerFunc(wsEchoHandler))

	// Convert http://127.0.0.1 to ws://127.0.0.
	var err error
	rawURL, err := url.Parse(wsts.server.URL)
	if err != nil {
		return nil, err
	}
	rawURL.Scheme = "ws"
	wsts.url = *rawURL

	return &wsts, nil
}

// defer this
func (wsts *wsTestServer) Close() {
	wsts.server.CloseClientConnections()
	wsts.server.Close()
}

/*
func (wst *wsTester) Start() error {
	// Connect to the server
	var err error
	wst.ws, _, err = websocket.DefaultDialer.Dial(wst.url.String(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (wst *wsTester) Stop() error {
	wst.ws.Close()
	wst.ws = nil

	return nil
}
*/

func wsEchoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("wsEchoHandler")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		fmt.Printf("recv: %s\n", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

/*
func TestExample(t *testing.T) {
	require := require.New(t)

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	rawURL, err := url.Parse(s.URL)
	require.NoError(err)

	rawURL.Scheme = "ws"

	//	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}
*/
