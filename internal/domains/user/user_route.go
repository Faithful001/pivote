package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewUserController()

	router.POST("", controller.CreateUser)
	router.GET("", controller.GetAllUsers)
	router.GET("/:id", controller.GetUser)
	router.PUT("/:id", controller.UpdateUser)
	router.DELETE("/:id", controller.DeleteUser)
}
