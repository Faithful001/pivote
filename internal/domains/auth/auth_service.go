package auth

import (
	"errors"
	"os"

	authdto "pivote/internal/domains/auth/dto"
	"pivote/internal/domains/otp"
	otpdto "pivote/internal/domains/otp/dto"
	"pivote/internal/domains/user"
	"pivote/internal/infra/db"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/types"
	"pivote/internal/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	userService *user.UserService
	otpService  *otp.OtpService
	jwtSecret string
}

func NewAuthService(mq *rabbitmq.RabbitMQ) *AuthService {
	return &AuthService{
		userService: user.NewUserService(),
		otpService:  otp.NewOtpService(mq),
		jwtSecret: os.Getenv("JWT_SECRET"),
	}
}


type AuthResponse struct {
	User  		*user.User `json:"user"`
	AccessToken string     `json:"token"`
}

func (s *AuthService) Register(payload authdto.RegisterDto) (*user.User, error) {
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

	err = s.otpService.SendOtpToEmail(userCreated.Email, otpdto.PurposeVerifyAcct)
	if err != nil {
		return nil, err
	}

	return userCreated, nil
}

func (s *AuthService) Login(email, password string) (*AuthResponse, error) {
	// Find user by email
	user, err := s.userService.GetUserByEmail(email)
	
	if err != nil {
		if err.Error() == "user not found" || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Invalid email or password")
		}
		return nil, err
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, errors.New("Invalid email or password")
	}

	if user.IsVerified == false {
		return nil, errors.New("user not verified")
	}

	// Generate JWT token
	
	jwtUtil, err := utils.NewJWTUtil(s.jwtSecret)

	if err != nil {
		return nil, err
	}


	token, err := jwtUtil.GenerateToken(utils.TokenOptions{
		UserID:    user.ID,
		Role:      string(user.Role),
		Purpose:   types.JwtPurposeAuthentication,
	})
	
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:  user,
		AccessToken: token,
	}, nil
}

func (s *AuthService) VerifyAccount(email string, otpString string) (*user.User, error) {
	// Verify the OTP
	if err := s.otpService.VerifyOtp(email, otpString, otpdto.PurposeVerifyAcct); err != nil {
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

