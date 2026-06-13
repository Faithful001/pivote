package workspace

import (
	"pivote/internal/domains/user"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	controller := NewWorkspaceController()

	workspaceGroup := router.Group("/workspaces")
	workspaceGroup.Use(middlewares.Standard(user.RoleAdmin))

	workspaceGroup.POST("", controller.CreateWorkspace)
	workspaceGroup.GET("", controller.GetWorkspaces)
	workspaceGroup.GET("/:id", controller.GetWorkspaceId)
	workspaceGroup.PUT("/:id", controller.UpdateWorkspace)
	workspaceGroup.DELETE("/:id", controller.DeleteWorkspace)
}
