package user

import "github.com/gin-gonic/gin"

// RegisterRoutes registers user routes onto two pre-configured groups.
//
//   - userGroup  → authenticated (admin + user) — me, update
//   - adminGroup → admin-only — list, get by id, delete
//
// Middleware is applied by the caller (router.go) to avoid an import cycle,
// since the middlewares package already imports the user package.
func RegisterRoutes(userGroup *gin.RouterGroup, adminGroup *gin.RouterGroup) {
	controller := NewUserController()

	// Any authenticated user
	userGroup.GET("/me", controller.GetMe)      // GET    /users/me
	userGroup.PUT("/:id", controller.UpdateUser) // PUT    /users/:id

	// Admin only
	adminGroup.GET("", controller.GetAllUsers)   // GET    /users
	adminGroup.GET("/:id", controller.GetUser)   // GET    /users/:id
	adminGroup.DELETE("/:id", controller.DeleteUser) // DELETE /users/:id
}
