package candidate

import (
	"pivote/internal/domains/user"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all candidate-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewCandidateController()

	// Public read routes (any authenticated user can browse candidates)
	protected := router.Group("/")
	protected.Use(middlewares.Standard(user.RoleAdmin, user.RoleUser))

	protected.GET("", controller.GetCandidates)                           // GET  /candidates
	protected.GET("/:id", controller.GetCandidateById)                   // GET  /candidates/:id
	protected.GET("/program/:program_id", controller.GetCandidatesByProgramID) // GET  /candidates/program/:program_id

	// Admin-only write routes
	admin := router.Group("/")
	admin.Use(middlewares.Standard(user.RoleAdmin))

	admin.POST("", controller.CreateCandidate)       // POST   /candidates
	admin.PUT("/:id", controller.UpdateCandidate)    // PUT    /candidates/:id
	admin.DELETE("/:id", controller.DeleteCandidate) // DELETE /candidates/:id
}