package websocket

import (
	"log"
	"net/http"
	"os"
	"pivote/internal/infra/db"
	"pivote/internal/utils"

	"github.com/google/uuid"
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

type JoinPayload struct {
	Token       string `json:"token"`
	ProgramID   string `json:"program_id"`
	WorkspaceID string `json:"workspace_id"`
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

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtUtil, err := utils.NewJWTUtil(jwtSecret)
	if err != nil {
		log.Printf("Socket.IO: JWT utility initialization failed: %v", err)
	}

	// Connection event
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("Socket.IO client connected:", s.ID())
		return nil
	})

	// Join program room event
	server.OnEvent("/", "join", func(s socketio.Conn, payload JoinPayload) {
		if jwtUtil == nil {
			s.Emit("error", map[string]string{"message": "JWT utility not initialized"})
			return
		}

		claims, err := jwtUtil.ParseToken(payload.Token)
		if err != nil {
			s.Emit("error", map[string]string{"message": "Unauthorized: invalid token"})
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			s.Emit("error", map[string]string{"message": "Unauthorized: invalid user ID"})
			return
		}

		programID, err := uuid.Parse(payload.ProgramID)
		if err != nil {
			s.Emit("error", map[string]string{"message": "Invalid program ID"})
			return
		}

		authorized := false
		if claims.Role == "admin" {
			// Admins check program ownership
			var count int64
			db.DB.Table("programs").Where("id = ? AND owner_id = ?", programID, userID).Count(&count)
			if count > 0 {
				authorized = true
			}
		} else {
			// Standard users check user_programs table
			var count int64
			db.DB.Table("user_programs").Where("user_id = ? AND program_id = ?", userID, programID).Count(&count)
			if count > 0 {
				authorized = true
			}
		}

		if !authorized {
			s.Emit("error", map[string]string{"message": "Unauthorized: you do not belong to this program"})
			return
		}

		// Leave all other program rooms first
		for _, room := range s.Rooms() {
			if room != s.ID() && room != "/" {
				s.Leave(room)
			}
		}

		s.Join("program:" + payload.ProgramID)
		log.Printf("Socket.IO client %s joined room program:%s", s.ID(), payload.ProgramID)
		s.Emit("joined", map[string]string{"program_id": payload.ProgramID})
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

// BroadcastLeaderboard broadcasts the leaderboard update event to clients in the program room.
func (s *SocketIOServer) BroadcastLeaderboard(programID string, data interface{}) {
	s.Server.BroadcastToRoom("/", "program:"+programID, "leaderboard:update", data)
}

// BroadcastVote broadcasts the vote event to clients in the program room.
func (s *SocketIOServer) BroadcastVote(programID string, data interface{}) {
	s.Server.BroadcastToRoom("/", "program:"+programID, "vote:broadcast", data)
}

// BroadcastStartProgram broadcasts the start program event to clients in the program room.
func (s *SocketIOServer) BroadcastStartProgram(programID string, data interface{}) {
	s.Server.BroadcastToRoom("/", "program:"+programID, "program:start", data)
}

// BroadcastStopProgram broadcasts the stop program event to clients in the program room.
func (s *SocketIOServer) BroadcastStopProgram(programID string, data interface{}) {
	s.Server.BroadcastToRoom("/", "program:"+programID, "program:stop", data)
}