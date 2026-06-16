package auth

import (
	"net/http"
	"pivote/internal/infra/rabbitmq"

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

	userCreated, err := ctrl.service.Register(payload, false)
	
	if err != nil {
		status := http.StatusInternalServerError

		if err.Error() == "Email already in use" {
			status = http.StatusConflict
		}
		
		c.JSON(status, gin.H{
			"statusCode": status,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
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
	userLogged, err := ctrl.service.Login(credentials.Email, credentials.Password, false)
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
	
	userVerified, err := ctrl.service.VerifyAccount(payload.Email, payload.Otp, false)
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

func (ctrl *AuthController) ForgotPassword(c *gin.Context) {
	var payload dto.ForgotPasswordDto

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	err := ctrl.service.ForgotPassword(payload.Email, false)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"statusCode": status,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "OTP sent successfully",
		"data":       nil,
	})
}

func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var payload dto.ResetPasswordDto

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	err := ctrl.service.ResetPassword(payload.Email, payload.Otp, payload.Password, false)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"statusCode": status,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Password reset successfully",
		"data":       nil,
	})
}

func (ctrl *AuthController) AdminRegister(c *gin.Context) {
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

	userCreated, err := ctrl.service.Register(payload, true)
	
	if err != nil {
		status := http.StatusInternalServerError

		if err.Error() == "Email already in use" {
			status = http.StatusConflict
		}
		
		c.JSON(status, gin.H{
			"statusCode": status,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
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

func (ctrl *AuthController) AdminLogin(c *gin.Context) {

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
	userLogged, err := ctrl.service.Login(credentials.Email, credentials.Password, true)
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

func (ctrl *AuthController) AdminVerifyAccount (c *gin.Context){
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
	
	userVerified, err := ctrl.service.VerifyAccount(payload.Email, payload.Otp, true)
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

func (ctrl *AuthController) AdminForgotPassword(c *gin.Context) {
	var payload dto.ForgotPasswordDto

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	err := ctrl.service.ForgotPassword(payload.Email, true)
	if err != nil {
		status := http.StatusInternalServerError

		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"statusCode": status,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "OTP sent successfully",
		"data":       nil,
	})
}

func (ctrl *AuthController) AdminResetPassword(c *gin.Context) {
	var payload dto.ResetPasswordDto

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	err := ctrl.service.ResetPassword(payload.Email, payload.Otp, payload.Password, true)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"statusCode": status,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Password reset successfully",
		"data":       nil,
	})
}