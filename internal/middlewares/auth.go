package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"pivote/internal/domains/user"
	"pivote/internal/infra/db"
	"pivote/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const AuthUserKey contextKey = "auth_user"

type AuthConfig struct {
	SecretKey     string
	RequiredRoles []user.Role
	RequiredPerms []string
	SkipIfNoToken bool
	TokenLookup   string
}

func Auth(cfg AuthConfig) gin.HandlerFunc {
	if cfg.SecretKey == "" {
		panic("JWT Secret Key is required")
	}

	return func(c *gin.Context) {
		tokenStr, err := extractToken(c, cfg.TokenLookup)
		if err != nil {
			if cfg.SkipIfNoToken {
				c.Next()
				return
			}
			unauthorized(c, "Token not found")
			return
		}

		claims := &types.CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.SecretKey), nil
		})

		if err != nil || !token.Valid {
			unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Parse user ID from claims
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			unauthorized(c, "Invalid user ID in token")
			c.Abort()
			return
		}

		var authUser user.User

		result := db.DB.Where("id = ?", userID).First(&authUser)

		if result.Error != nil {
			unauthorized(c, "Invalid user ID in token")
			c.Abort()
			return
		}

		// Construct user object from claims
		// authUser = user.User{
		// 	ID:    		userID,
		// 	Name:  		claims.Name,
		// 	Email: 		claims.Email,
		// 	Role:  		user.Role(claims.Role),
		// 	IsVerified: claims.IsVerified,
		// }

		//IsVerified check
		if !authUser.IsVerified {
			forbidden(c, "User is not verified")
			c.Abort()
			return
		}

		//purpose check
		if claims.Purpose != types.JwtPurposeAuthentication {
			forbidden(c, "Invalid token purpose")
			c.Abort()
			return
		}

		// Role check
		if len(cfg.RequiredRoles) > 0 && !hasAnyRole(authUser.Role, cfg.RequiredRoles) {
			forbidden(c, "Insufficient role")
			c.Abort()
			return
		}

		// Permission check
		if len(cfg.RequiredPerms) > 0 && !hasAllPermissions(claims.Permissions, cfg.RequiredPerms) {
			forbidden(c, "Missing required permissions")
			c.Abort()
			return
		}

		// Save user in context
		c.Set(string(AuthUserKey), authUser)
		c.Next()
	}
}

// Standard returns an Auth middleware with default configuration.
// It automatically loads the secret from JWT_SECRET env var.
// You can pass required roles as arguments.
func Standard(requiredRoles ...user.Role) gin.HandlerFunc {
	return Auth(AuthConfig{
		SecretKey:     os.Getenv("JWT_SECRET"),
		RequiredRoles: requiredRoles,
		TokenLookup:   "header:Authorization:Bearer,cookie:access_token,query:token",
	})
}

func extractToken(c *gin.Context, lookup string) (string, error) {
	if lookup == "" {
		lookup = "header:Authorization:Bearer,cookie:access_token"
	}

	for _, source := range strings.Split(lookup, ",") {
		parts := strings.SplitN(strings.TrimSpace(source), ":", 3)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "header":
			auth := c.GetHeader(parts[1])
			if len(parts) == 3 && strings.HasPrefix(auth, parts[2]+" ") {
				return strings.TrimPrefix(auth, parts[2]+" "), nil
			}
		case "cookie":
			if token, err := c.Cookie(parts[1]); err == nil && token != "" {
				return token, nil
			}
		case "query":
			if token := c.Query(parts[1]); token != "" {
				return token, nil
			}
		}
	}
	return "", errors.New("token not found")
}

// Role & Permission checks
func hasAnyRole(userRole user.Role, allowed []user.Role) bool {
	for _, r := range allowed {
		if userRole == r {
			return true
		}
	}
	return false
}

func hasAllPermissions(userPerms, required []string) bool {
	permSet := make(map[string]bool)
	for _, p := range userPerms {
		permSet[p] = true
	}
	for _, p := range required {
		if !permSet[p] {
			return false
		}
	}
	return true
}

// Response helpers
func unauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + msg})
}

func forbidden(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: " + msg})
}

// GetUser extracts the strongly-typed user from the context
func GetUser(c *gin.Context) (user.User, error) {
	val, exists := c.Get(string(AuthUserKey))
	if !exists {
		return user.User{}, errors.New("user not found in context")
	}
	// Assert type
	u, ok := val.(user.User)
	if !ok {
		return user.User{}, errors.New("invalid user type in context")
	}
	return u, nil
}