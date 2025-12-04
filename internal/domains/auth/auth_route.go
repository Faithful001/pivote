package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all auth-related routes
func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewAuthController()

	// Auth routes
	router.POST("/register", controller.Register) // POST /auth/register
	router.POST("/login", controller.Login)       // POST /auth/login
}
