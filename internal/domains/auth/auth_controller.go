package auth

import (
	"net/http"
	"pivote/internal/infra/rabbitmq"
	"strconv"

	"pivote/internal/domains/auth/dto"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *AuthService
}

func NewAuthController(mq *rabbitmq.RabbitMQ) *AuthController {
	return &AuthController{
		service: NewAuthService(mq),
	}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var payload dto.RegisterDto

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	userCreated, err := ctrl.service.Register(payload)
	if err != nil {
		// Check if it's a duplicate email error
		if err.Error() == "User with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"statusCode": http.StatusConflict,
				"success":    false,
				"message":    err.Error(),
				"data":       nil,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":    false,
			"message":    "Failed to create user",
			"data":       nil,
			"error":      err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"success":    true,
		"message":    "User created successfully",
		"data":       userCreated,
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {

	var credentials dto.LoginDto

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	// Call service layer to authenticate
	userLogged, err := ctrl.service.Login(credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "User logged in successfully",
		"data":       userLogged,
	})
}

func (ctrl *AuthController) VerifyAccount (c *gin.Context){
	var payload dto.VerifyAccountDto
	
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}
	
	userVerified, err := ctrl.service.VerifyAccount(payload.Email, strconv.Itoa(payload.Otp))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "User verified successfully",
		"data":       userVerified,
	})
}