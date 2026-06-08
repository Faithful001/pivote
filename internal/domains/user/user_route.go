package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(userGroup *gin.RouterGroup, adminGroup *gin.RouterGroup) {
	controller := NewUserController()

	userGroup.GET("/me", controller.GetMe)      
	userGroup.PUT("/:id", controller.UpdateUser) 

	adminGroup.GET("", controller.GetAllUsers)   
	adminGroup.GET("/:id", controller.GetUser)   
	adminGroup.DELETE("/:id", controller.DeleteUser) 
}
