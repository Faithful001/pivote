package dto

type Purpose string

const (
	PurposeVerifyAcct      Purpose = "verify_account"
	PurposeResetPwd        Purpose = "reset_pwd"
	PurposeRequestJoinLink Purpose = "request_join_link"
	PurposeRegisterToJoin  Purpose = "register_to_join"
)

type SendOtpDto struct {
	Email   string  `json:"email" binding:"required,email"`
	Purpose Purpose `json:"purpose" binding:"required"`
}