package program

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all program-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewProgramController()

	// Program routes
	router.POST("/programs", controller.CreateProgram)       // POST /programs
	router.GET("/programs", controller.GetPrograms)          // GET /programs
	router.GET("/programs/:id", controller.GetProgramById)   // GET /programs/:id
	router.PUT("/programs/:id", controller.UpdateProgram)    // PUT /programs/:id
	router.DELETE("/programs/:id", controller.DeleteProgram) // DELETE /programs/:id
}
