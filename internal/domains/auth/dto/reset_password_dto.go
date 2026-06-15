package dto

type ResetPasswordDto struct {
	Email    string `json:"email" binding:"required,email"`
	Otp      string `json:"otp" binding:"required,len=4"`
	Password string `json:"password" binding:"required,min=6"`
}
