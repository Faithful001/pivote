package types

import "github.com/golang-jwt/jwt/v5"

// CustomClaims represents the custom JWT claims with additional fields
type CustomClaims struct {
	jwt.RegisteredClaims
	UserID      string   	`json:"user_id"`
	Email       string   	`json:"email"`
	Name        string		`json:"name"`
	IsVerified	bool		`json:"is_verified"`
	Role        string   	`json:"role"`
	Permissions []string 	`json:"permissions,omitempty"`
}
