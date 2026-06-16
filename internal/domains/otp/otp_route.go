package otp

import (
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ) {
	controller := NewOtpController(mq)

	router.POST("/send",
		middlewares.RateLimitByIP("otp:send", 5, 10*time.Minute),
		controller.SendOtp,
	)

	router.POST("/verify",
		middlewares.RateLimitByIP("otp:verify", 10, 10*time.Minute),
		controller.VerifyOtp,
	)
}
