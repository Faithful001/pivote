package program

import (
	"fmt"
	"net/http"
	dto "pivote/internal/domains/program/dto"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/infra/sse"
	"pivote/internal/middlewares"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProgramController struct {
	service        *ProgramService
	mq             *rabbitmq.RabbitMQ
	sseBroadcaster *sse.BroadcasterManager
}

func NewProgramController(mq *rabbitmq.RabbitMQ, sseBroadcaster *sse.BroadcasterManager) (*ProgramController, error) {
	programService, err := NewProgramService(mq)
	if err != nil {
		return nil, err
	}

	return &ProgramController{
		service:        programService,
		mq:             mq,
		sseBroadcaster: sseBroadcaster,
	}, nil
}

func (ctrl *ProgramController) CreateProgram(c *gin.Context) {
	user, err := middlewares.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
		return
	}

	var payload dto.CreateProgramDto

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.CreateProgram(user.ID, payload)

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
	// Get user from context to check program enrollment
	user, err := middlewares.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
		return
	}
	userID := user.ID
	
	//workspace_id
	workspaceId := c.Query("workspace_id")

	//parse the workspaceId to a uuid only if it's provided
	var workspaceID uuid.UUID
	if workspaceId != "" {
		var err error
		workspaceID, err = uuid.Parse(workspaceId)
		if err != nil && user.Role == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{
				"statusCode": http.StatusBadRequest,
				"success":    false,
				"message":    "Invalid workspace id",
				"data":       nil,
			})
			return
		}
	}


	result, err := ctrl.service.GetPrograms(userID, workspaceID, user.Role)

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

	// //workspace_id
	// workspaceId := c.Query("workspace_id")

	// //parse the workspaceId to a uuid
	// workspaceID, err := uuid.Parse(workspaceId)

	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"statusCode": http.StatusBadRequest,
	// 		"success":    false,
	// 		"message":    "Invalid workspace id",
	// 		"data":       nil,
	// 	})
	// 	return
	// }

	user, err := middlewares.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
		return
	}
	userID := user.ID

	result, err := ctrl.service.GetProgramById(id, userID, user.Role)

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

	var payload dto.UpdateProgramDto

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

func (ctrl *ProgramController) JoinProgram(c *gin.Context) {
	idParam := c.Param("id")
	programID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid program ID",
			"data":       nil,
		})
		return
	}

	var payload struct {
		// Email string `json:"email" binding:"required"`
		WorkspaceID	string	`json:"workspace_id" binding:"required"`
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Workspace id and token are required",
			"data":       nil,
		})
		return
	}

	workspaceID, err := uuid.Parse(payload.WorkspaceID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid id",
			"data":       nil,
		})
		return
	}

	err = ctrl.service.JoinProgram(programID, workspaceID, payload.Token)

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
		"message":    "Successfully joined program",
		"data":       nil,
	})
}

func (ctrl *ProgramController) ToggleProgram(c *gin.Context) {
	idParam := c.Param("id")
	programID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid program ID",
			"data":       nil,
		})
		return
	}

	var payload struct {
		IsActive 		bool `json:"is_active"`
		VotingEndsAt 	string `json:"voting_ends_at"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.ToggleProgram(programID, payload.IsActive, payload.VotingEndsAt)
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
		"message":    fmt.Sprintf("Voting status %s", map[bool]string{
			true:  "turned on",
			false: "turned off",
		}[result.IsActive]),
		"data": result,
	})
}

func (ctrl *ProgramController) RequestJoinLink(c *gin.Context) {
	idParam := c.Param("id")
	programID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid program ID",
			"data":       nil,
		})
		return
	}

	var payload struct {
		Email string `json:"email" binding:"required"`
		WorkspaceID	string	`json:"workspace_id" binding:"required"`
	}

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	workspaceID, err := uuid.Parse(payload.WorkspaceID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid id",
			"data":       nil,
		})
		return
	}

	err = ctrl.service.RequestJoinLink(payload.Email, programID, workspaceID)
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
		"message":    "Join link sent to email successfully",
		"data":       nil,
	})
}

func (ctrl *ProgramController) StreamCountdown(c *gin.Context) {
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

	user, err := middlewares.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    "Unauthorized",
			"data":       nil,
		})
		return
	}

	// Verify program exists and user has access to it
	result, err := ctrl.service.GetProgramById(id, user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"statusCode": http.StatusNotFound,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	if result.VotingEndsAt == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "voting_ends_at is not set for this program",
			"data":       nil,
		})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Send initial countdown value immediately
	remaining := int64(time.Until(*result.VotingEndsAt).Seconds())
	if remaining < 0 || !result.IsActive {
		remaining = 0
	}
	c.SSEvent("countdown", remaining)
	c.Writer.Flush()

	// If already expired or inactive, we can stop here
	if remaining <= 0 {
		return
	}

	// Register client to the broadcaster
	ch := ctrl.sseBroadcaster.Register(id, *result.VotingEndsAt, result.IsActive)
	defer ctrl.sseBroadcaster.Unregister(id, ch)

	// Stream countdown events
	for {
		select {
		case rem, ok := <-ch:
			if !ok {
				// Ticker stopped or program expired
				c.SSEvent("countdown", 0)
				c.Writer.Flush()
				return
			}
			c.SSEvent("countdown", rem)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			// Client disconnected
			return
		}
	}
}