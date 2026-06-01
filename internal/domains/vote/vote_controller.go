package vote

import (
	"net/http"
	"pivote/internal/domains/vote/dtos"
	"pivote/internal/infra/websocket"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VoteController struct {
	service *VoteService
}

func NewVoteController(hub *websocket.Hub) VoteController {
	return VoteController{
		service: NewVoteService(hub),
	}
}

func (ctrl *VoteController) ToggleVote(c *gin.Context) {
	var payload dtos.ToggleVoteDto

	// Get user from context
	user, err := middlewares.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := user.ID 

	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request payload",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.ToggleVoteCandidate(userID, payload.CandidateID)

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
		"message":    "Vote toggled successfully",
		"data":       result,
	})
}

func (ctrl *VoteController) GetVotesByProgramID(c *gin.Context) {
	idParam := c.Param("program_id")
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

	result, err := ctrl.service.GetVotesByProgramID(programID)

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
		"message":    "Votes retrieved successfully",
		"data":       result,
	})
}

func (ctrl *VoteController) GetVotesByCandidateID(c *gin.Context) {
	idParam := c.Param("candidate_id")
	candidateID, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid candidate ID",
			"data":       nil,
		})
		return
	}

	result, err := ctrl.service.GetVotesByCandidateID(candidateID)

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
		"message":    "Votes retrieved successfully",
		"data":       result,
	})
}
