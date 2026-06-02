package websocket

import (
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

// SocketIOServer wraps the socket.io server to make it easy to inject and use in other packages (like vote service)
type SocketIOServer struct {
	Server *socketio.Server
}

// NewSocketIOServer initializes and configures the Socket.IO server.
// It sets up CORS options and registers connection/disconnection and message handlers.
func NewSocketIOServer() (*SocketIOServer, error) {
	// Configure CORS and transport options so client connections don't get rejected
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all origins for development. Adjust for production.
				},
			},
			&polling.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all origins for development. Adjust for production.
				},
			},
		},
	})

	// Connection event
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("Socket.IO client connected:", s.ID())
		return nil
	})

	// Disconnect event
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("Socket.IO client disconnected:", s.ID(), reason)
	})

	// Error event
	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("Socket.IO error:", e)
	})

	// Ping event
	server.OnEvent("/", "ping", func(s socketio.Conn) {
		log.Println("Socket.IO ping received")
		s.Emit("pong", map[string]string{
			"type": "pong",
		})
	})

	// Generic message event example
	server.OnEvent("/", "message", func(s socketio.Conn, msg string) {
		log.Println("Socket.IO message:", msg)
		server.BroadcastToNamespace("/", "message", msg)
	})

	// Start the background server handler
	go func() {
		if err := server.Serve(); err != nil {
			log.Printf("Socket.IO serve error: %v", err)
		}
	}()

	return &SocketIOServer{Server: server}, nil
}

// BroadcastLeaderboard broadcasts the leaderboard update event to all connected clients.
func (s *SocketIOServer) BroadcastLeaderboard(data interface{}) {
	s.Server.BroadcastToNamespace("/", "leaderboard:update", data)
}