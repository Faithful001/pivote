package utils

import (
	"errors"
	"pivote/internal/types"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenOptions struct {
	UserID    uuid.UUID
	Role      string
	Purpose   types.JwtPurpose
	ProgramID *uuid.UUID
	WorkspaceID *uuid.UUID
	ExpiresAt *time.Time // nil defaults to 24h
}

type JWTUtil struct {
	secret []byte
}

func NewJWTUtil(secret string) (*JWTUtil, error) {
	if secret == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	return &JWTUtil{secret: []byte(secret)}, nil
}

func (j *JWTUtil) GenerateToken(opts TokenOptions) (string, error) {
	expiry := time.Now().Add(24 * time.Hour)
	if opts.ExpiresAt != nil {
		expiry = *opts.ExpiresAt
	}

	claims := types.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   opts.UserID.String(),
		},
		Role:      opts.Role,
		Purpose:   opts.Purpose,
		ProgramID: opts.ProgramID,
		WorkspaceID: opts.WorkspaceID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTUtil) ParseToken(tokenStr string) (*types.CustomClaims, error) {
	claims := &types.CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}