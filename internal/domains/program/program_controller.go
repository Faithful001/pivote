package program

import (
	"net/http"
	"pivote/internal/domains/program/dtos"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProgramController struct {
	service *ProgramService
}

func NewProgramController() ProgramController {
	return ProgramController{
		service: NewProgramService(),
	}
}

func (ctrl *ProgramController) CreateProgram(c *gin.Context) {
	var payload dtos.CreateProgramDto

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.CreateProgram(payload)

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
		"message":    "Program created successfully",
		"data":       result,
	})
}

func (ctrl *ProgramController) GetPrograms(c *gin.Context) {
	result, err := ctrl.service.GetPrograms()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Programs retrieved successfully",
		"data":       result,
	})
}

func (ctrl *ProgramController) GetProgramById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid program ID",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.GetProgramById(id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"statusCode": http.StatusNotFound,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Program retrieved successfully",
		"data":       result,
	})
}

func (ctrl *ProgramController) UpdateProgram(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid program ID",
			"data":       nil,
		})
		return
	}

	var payload dtos.UpdateProgramDto

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.UpdateProgram(id, payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Program updated successfully",
		"data":       result,
	})
}

func (ctrl *ProgramController) DeleteProgram(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid program ID",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.DeleteProgram(id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"statusCode": http.StatusNotFound,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Program deleted successfully",
		"data":       result,
	})
}