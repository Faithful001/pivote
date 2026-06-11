package program

import (
	"fmt"
	"pivote/internal/domains/user"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/infra/sse"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all program-related routes
func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ, sseBroadcaster *sse.BroadcasterManager) {
	controller, err := NewProgramController(mq, sseBroadcaster)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize program controller: %v", err))
	}

	router.POST("/:id/join", controller.JoinProgram) 
	router.POST("/:id/request-join", controller.RequestJoinLink)

	// User & Admin protected routes
	protected := router.Group("")
	protected.Use(middlewares.Standard(user.RoleAdmin, user.RoleUser))

	protected.GET("", controller.GetPrograms) 
	protected.GET("/:id", controller.GetProgramById)       
	protected.GET("/:id/countdown", controller.StreamCountdown)

	// Admin-only write routes
	admin := router.Group("")
	admin.Use(middlewares.Standard(user.RoleAdmin))

	admin.POST("", controller.CreateProgram)             // POST /programs
	admin.PUT("/:id", controller.UpdateProgram)          // PUT /programs/:id
	admin.DELETE("/:id", controller.DeleteProgram)       // DELETE /programs/:id
	admin.PATCH("/:id/toggle", controller.ToggleProgram) // PATCH /programs/:id/toggle
}
