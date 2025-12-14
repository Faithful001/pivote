package dto

type Purpose string

const (
	PurposeVerifyAcct Purpose = "verify_acct"
	PurposeResetPwd   Purpose = "reset_pwd"
)

type SendOtpDto struct {
	Email   string  `json:"email" binding:"required,email"`
	Purpose Purpose `json:"purpose" binding:"required"`
}