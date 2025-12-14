package candidate

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all candidate-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewCandidateController()

	// Candidate routes
	router.POST("/candidates", controller.CreateCandidate)       // POST /candidates
	router.GET("/candidates", controller.GetCandidates)          // GET /candidates
	router.GET("/candidates/:id", controller.GetCandidateById)   // GET /candidates/:id
	router.PUT("/candidates/:id", controller.UpdateCandidate)    // PUT /candidates/:id
	router.DELETE("/candidates/:id", controller.DeleteCandidate) // DELETE /candidates/:id
}