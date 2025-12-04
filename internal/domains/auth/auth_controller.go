package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{
		service: NewAuthService(),
	}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var payload RegisterPayload

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

	var credentials struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

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
