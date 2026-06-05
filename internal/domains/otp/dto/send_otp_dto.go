package dto

type Purpose string

const (
	PurposeVerifyAcct Purpose = "verify_acct"
	PurposeResetPwd   Purpose = "reset_pwd"
	PurposeRequestVote Purpose = "request_vote"
)

type SendOtpDto struct {
	Email   string  `json:"email" binding:"required,email"`
	Purpose Purpose `json:"purpose" binding:"required"`
}