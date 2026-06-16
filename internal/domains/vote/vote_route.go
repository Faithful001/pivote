package vote

import (
	"pivote/internal/domains/user"
	"pivote/internal/infra/websocket"
	"pivote/internal/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all vote-related routes
func RegisterRoutes(router *gin.RouterGroup, socketio *websocket.SocketIOServer) {
	controller := NewVoteController(socketio)

	protected := router.Group("/")
	protected.Use(middlewares.Standard(user.RoleAdmin, user.RoleUser))

	protected.POST("/toggle",
		middlewares.RateLimit(middlewares.RateLimitConfig{
			KeyPrefix: "vote:toggle",
			Max:       20,
			Window:    time.Minute,
			UseUserID: true,
		}),
		controller.ToggleVote,
	)

	protected.GET("/program/:program_id", controller.GetVotesByProgramID)
	protected.GET("/candidate/:candidate_id", controller.GetVotesByCandidateID)
}
