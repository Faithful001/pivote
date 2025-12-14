package auth

import (
	"pivote/internal/infra/rabbitmq"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all auth-related routes
func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ) {
	controller := NewAuthController(mq)

	// Auth routes
	router.POST("/register", controller.Register) // POST /auth/register
	router.POST("/login", controller.Login)       // POST /auth/login
}
