package auth

import (
	"pivote/internal/infra/rabbitmq"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ) {
	controller := NewAuthController(mq)

	router.POST("/register", controller.Register) 
	router.POST("/login", controller.Login)       
	router.POST("/verify-account", controller.VerifyAccount)      
}
