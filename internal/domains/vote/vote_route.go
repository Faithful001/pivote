package vote

import (
	"pivote/internal/domains/user"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// RegisterRoutes registers all vote-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewVoteController()

	errRes := godotenv.Load()
	if errRes != nil {
		panic("Error loading .env file")
	}

	protected := router.Group("/")
	// Use Standard middleware (Roles: Admin, User)
	// Keeps it simple! If you need specific permissions (e.g. "read", "write"), you can use the full struct.
	protected.Use(middlewares.Standard(user.RoleAdmin, user.RoleUser))

	// Vote routes
	protected.POST("/votes/toggle", controller.ToggleVote)                                 // POST /votes/toggle
	protected.GET("/votes/program/:program_id", controller.GetVotesByProgramID)           // GET /votes/program/:program_id
	protected.GET("/votes/candidate/:candidate_id", controller.GetVotesByCandidateID)     // GET /votes/candidate/:candidate_id
}
