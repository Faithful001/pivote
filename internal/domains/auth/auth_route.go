package auth

import (
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, mq *rabbitmq.RabbitMQ) {
	controller := NewAuthController(mq)

	router.POST("/register",
		middlewares.RateLimitByIP("auth:register", 10, 10*time.Minute),
		controller.Register,
	)

	router.POST("/login",
		middlewares.RateLimitByIP("auth:login", 10, 5*time.Minute),
		controller.Login,
	)

	router.POST("/verify-account",
		middlewares.RateLimitByIP("auth:verify-account", 10, 10*time.Minute),
		controller.VerifyAccount,
	)

	router.POST("/forgot-password",
		middlewares.RateLimitByIP("auth:forgot-password", 5, 15*time.Minute),
		controller.ForgotPassword,
	)

	router.POST("/reset-password",
		middlewares.RateLimitByIP("auth:reset-password", 5, 15*time.Minute),
		controller.ResetPassword,
	)

	router.POST("/admin/register",
		middlewares.RateLimitByIP("auth:register-admin", 10, 10*time.Minute),
		controller.AdminRegister,
	)

	router.POST("/admin/login",
		middlewares.RateLimitByIP("auth:login-admin", 10, 5*time.Minute),
		controller.AdminLogin,
	)

	router.POST("/admin/verify-account",
		middlewares.RateLimitByIP("auth:verify-account-admin", 10, 10*time.Minute),
		controller.AdminVerifyAccount,
	)

	router.POST("/admin/forgot-password",
		middlewares.RateLimitByIP("auth:forgot-password-admin", 5, 15*time.Minute),
		controller.AdminForgotPassword,
	)

	router.POST("/admin/reset-password",
		middlewares.RateLimitByIP("auth:reset-password-admin", 5, 15*time.Minute),
		controller.AdminResetPassword,
	)
}
