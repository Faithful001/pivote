package program

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all program-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewProgramController()

	// Program routes
	router.POST("/", controller.CreateProgram)       // POST /programs
	router.GET("/", controller.GetPrograms)          // GET /programs
	router.GET("/:id", controller.GetProgramById)   // GET /programs/:id
	router.PUT("/:id", controller.UpdateProgram)    // PUT /programs/:id
	router.DELETE("/:id", controller.DeleteProgram) // DELETE /programs/:id
}
