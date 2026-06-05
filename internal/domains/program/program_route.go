package program

import (
	"pivote/internal/domains/user"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all program-related routes
func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ) {
	controller := NewProgramController(mq)

	// User & Admin read routes
	protected := router.Group("")
	protected.Use(middlewares.Standard(user.RoleAdmin, user.RoleUser))

	protected.GET("", controller.GetPrograms)        // GET /programs
	protected.GET("/:id", controller.GetProgramById) // GET /programs/:id
	protected.POST("/:id/join", controller.JoinProgram) // POST /programs/:id/join
	// protected.GET("/:id/access-code", controller.GetAccessCode) // GET /programs/:id/access-code

	// Admin-only write routes
	admin := router.Group("")
	admin.Use(middlewares.Standard(user.RoleAdmin))

	admin.POST("", controller.CreateProgram)       // POST /programs
	admin.PUT("/:id", controller.UpdateProgram)    // PUT /programs/:id
	admin.DELETE("/:id", controller.DeleteProgram) // DELETE /programs/:id
	admin.PATCH("/:id/toggle", controller.ToggleProgram) // PATCH /programs/:id/toggle
}
