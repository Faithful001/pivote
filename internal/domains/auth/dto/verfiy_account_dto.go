package dto

type VerifyAccountDto struct {
	Email string `json:"email" binding:"required,email"`
	Otp   int    `json:"otp" binding:"required"`
}