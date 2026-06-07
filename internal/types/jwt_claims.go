package types

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtPurpose string

const (
	JwtPurposeAuthentication JwtPurpose = "authentication"
	JwtPurposeResetPwd       JwtPurpose = "reset_pwd"
	JwtPurposeProgramJoin    JwtPurpose = "program_join"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Email       string     `json:"email,omitempty"`
	Name        string     `json:"name,omitempty"`
	IsVerified  bool       `json:"is_verified,omitempty"`
	Role        string     `json:"role,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	Purpose     JwtPurpose `json:"purpose"`
	ProgramID   *uuid.UUID `json:"program_id,omitempty"`
}