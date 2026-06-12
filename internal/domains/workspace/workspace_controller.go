package workspace

import (
	"net/http"
	"pivote/internal/domains/workspace/dto"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorkspaceController struct {
	service WorkspaceService
}

func NewWorkspaceController() WorkspaceController {
	return WorkspaceController{
		service: NewWorkspaceService(),
	}
}

// admin only
func (wc *WorkspaceController) CreateWorkspace(c *gin.Context) {
	user, err := middlewares.GetUser(c)
	
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
	} 

	var payload dto.CreateWorkspaceDto

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	result, err := wc.service.CreateWorkspace(user.ID, payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"success":    true,
		"message":    "Workspace created successfully",
		"data":       &result,
	})
}

// admin only
func (wc *WorkspaceController) GetWorkspaces(c *gin.Context) {
	user, err := middlewares.GetUser(c)
	
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
	}

	result, err := wc.service.GetWorkspaces(user.ID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "All workspaces",
		"data":       &result,
	})
}

// admin only
func (wc *WorkspaceController) GetWorkspaceId(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid workspace ID",
			"data":       nil,
		})
		return
	}

	user, err := middlewares.GetUser(c)
	
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
	}

	result, err := wc.service.GetWorkspace(user.ID, id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Workspace retrieved",
		"data":       &result,
	})
}

// admin only
func (wc *WorkspaceController) UpdateWorkspace(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid workspace ID",
			"data":       nil,
		})
		return
	}

	user, err := middlewares.GetUser(c)
	
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
	} 

	var payload dto.UpdateWorkspaceDto

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	if err := wc.service.UpdateWorkspace(user.ID, id, payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"success":    true,
		"message":    "Workspace updated successfully",
		"data":       nil,
	})
}

// admin only
func (wc *WorkspaceController) DeleteWorkspace(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid workspace ID",
			"data":       nil,
		})
		return
	}

	user, err := middlewares.GetUser(c)
	
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
	}

	if err := wc.service.DeleteWorkspace(user.ID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Workspace deleted successfully",
		"data":       nil,
	})
}