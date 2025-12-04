package auth

import (
	"errors"
	"pivote/internal/domains/user"
	"pivote/internal/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	userService *user.UserService
}

func NewAuthService() *AuthService {
	return &AuthService{
		userService: user.NewUserService(),
	}
}

type RegisterPayload struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register creates a new user
func (s *AuthService) Register(payload RegisterPayload) (*user.User, error) {
	newUser := user.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
		Role:     "user",
	}

	userCreated, err := s.userService.CreateUser(&newUser)
	if err != nil {
		return nil, err
	}

	return userCreated, nil
}

// Login authenticates a user with email and password
func (s *AuthService) Login(email, password string) (*user.User, error) {
	// Find user by email
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "user not found" || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}
