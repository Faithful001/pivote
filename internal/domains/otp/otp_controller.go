package otp

import (
	"net/http"
	"pivote/internal/domains/otp/dto"
	"pivote/internal/infra/rabbitmq"

	"github.com/gin-gonic/gin"
)

type OtpController struct {
	service *OtpService
}

func NewOtpController(mq *rabbitmq.RabbitMQ) *OtpController {
	return &OtpController{
		service: NewOtpService(mq),
	}
}

func (ctrl *OtpController) SendOtp(c *gin.Context) {
	var payload dto.SendOtpDto

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	if payload.Purpose != dto.PurposeVerifyAcct && payload.Purpose != dto.PurposeResetPwd {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid purpose",
			"data":       nil,
		})
		return
	}

	if err :=ctrl.service.SendOtpToEmail(payload.Email, payload.Purpose); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":    false,
			"message":    "Failed to send OTP",
			"data":       nil,
			"error":      err.Error(),
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

func (ctrl *OtpController) VerifyOtp (c *gin.Context) {

	var payload struct {
		Email 	string `json:"email" binding:"required,email"`
		Otp 	string `json:"otp" binding:"required"`
		Purpose dto.Purpose `json:"purpose" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid request body",
			"data":       nil,
		})
		return
	}

	if err := ctrl.service.VerifyOtp(payload.Email, payload.Otp, payload.Purpose); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":    false,
			"message":    "Failed to verify OTP",
			"data":       nil,
			"error":      err.Error(),
		})
		return
	}

	// token, err := utils.GenerateToken(user.ID, string(user.Role))

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "OTP verified successfully",
		"data":       nil,
	})
}