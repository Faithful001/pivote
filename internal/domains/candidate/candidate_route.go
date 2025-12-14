package candidate

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all candidate-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewCandidateController()

	// Candidate routes
	router.POST("/", controller.CreateCandidate)       // POST /candidates
	router.GET("/", controller.GetCandidates)          // GET /candidates
	router.GET("/:id", controller.GetCandidateById)   // GET /candidates/:id
	router.PUT("/:id", controller.UpdateCandidate)    // PUT /candidates/:id
	router.DELETE("/:id", controller.DeleteCandidate) // DELETE /candidates/:id
}