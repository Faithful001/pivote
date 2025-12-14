package otp

import (
	"pivote/internal/infra/rabbitmq"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ) {
	controller := NewOtpController(mq)

	router.POST("/send", controller.SendOtp)
	router.POST("/verify", controller.VerifyOtp)
}
