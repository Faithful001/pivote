package auth

import (
	"errors"

	"pivote/internal/db"
	"pivote/internal/domains/otp"
	"pivote/internal/domains/user"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	userService *user.UserService
	otpService  *otp.OtpService
}

func NewAuthService(mq *rabbitmq.RabbitMQ) *AuthService {
	return &AuthService{
		userService: user.NewUserService(),
		otpService:  otp.NewOtpService(mq),
	}
}

type RegisterPayload struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

func (s *AuthService) Register(payload RegisterPayload) (*AuthResponse, error) {
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

	err = s.otpService.SendOtpToEmail(userCreated.Email, otp.PurposeVerifyAcct)
	if err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.GenerateToken(userCreated)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:  userCreated,
		Token: token,
	}, nil
}

func (s *AuthService) Login(email, password string) (*AuthResponse, error) {
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

	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *AuthService) VerifyAccount(email string, otpString string) (*user.User, error) {
	// Verify the OTP
	if err := s.otpService.VerifyOtp(email, otpString, otp.PurposeVerifyAcct); err != nil {
		return nil, err
	}

	// Update user verification status
	var userRecord user.User
	result := db.DB.Model(&user.User{}).Where("email = ?", email).Update("is_verified", true)
	
	if result.Error != nil {
		return nil, errors.New("failed to update user verification status")
	}

	// Fetch the updated user
	if err := db.DB.Where("email = ?", email).First(&userRecord).Error; err != nil {
		return nil, errors.New("failed to fetch user")
	}

	return &userRecord, nil
}

// GenerateToken creates a signed JWT token for a user
func (s *AuthService) GenerateToken(user *user.User) (string, error) {
	return utils.GenerateToken(user.ID, user.Email, user.Name, string(user.Role))
}
