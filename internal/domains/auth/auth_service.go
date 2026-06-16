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

func (s *AuthService) Register(payload authdto.RegisterDto, isAdmin bool) (*user.User, error) {
	var role user.Role
	if isAdmin {
		role = user.RoleAdmin
	} else {
		role = user.RoleUser
	}
	
	newUser := user.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
		Role:     role,
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

func (s *AuthService) Login(email, password string, isAdmin bool) (*AuthResponse, error) {
	var role user.Role
	if isAdmin {
		role = user.RoleAdmin
	} else {
		role = user.RoleUser
	}
	// Find user by email
	user, err := s.userService.GetUserByEmail(email)
	
	if err != nil {
		if err.Error() == "user not found" || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Invalid email or password")
		}
		return nil, err
	}

	// Check if user role matches
	if user.Role != role {
		return nil, errors.New("Invalid email or password")
	}

	if user.IsVerified == false {
		return nil, errors.New("user not verified")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, errors.New("Invalid email or password")
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

func (s *AuthService) VerifyAccount(email string, otpString string, isAdmin bool) (*user.User, error) {
	var role user.Role
	if isAdmin {
		role = user.RoleAdmin
	} else {
		role = user.RoleUser
	}
	// Find user by email
	foundUser, err := s.userService.GetUserByEmail(email)
	
	if err != nil {
		if err.Error() == "user not found" || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Invalid email or password")
		}
		return nil, err
	}

	// Check if user role matches
	if foundUser.Role != role {
		return nil, errors.New("Invalid email or password")
	}

	if foundUser.IsVerified == true {
		return nil, errors.New("User already verified")
	}
	
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

func (s *AuthService) ForgotPassword(email string, isAdmin bool) error {
	var role user.Role
	if isAdmin {
		role = user.RoleAdmin
	} else {
		role = user.RoleUser
	}
	// Find user by email
	foundUser, err := s.userService.GetUserByEmail(email)
	
	
	if err != nil {
		if err.Error() == "user not found" || errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Invalid email or password")
		}
		return err
	}

	// Check if user role matches
	if foundUser.Role != role {
		return errors.New("Invalid email or password")
	}


	// Generate and send OTP via otpService
	err = s.otpService.SendOtpToEmail(email, otpdto.PurposeResetPwd)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ResetPassword(email, otpString, newPassword string, isAdmin bool) error {
	var role user.Role
	if isAdmin {
		role = user.RoleAdmin
	} else {
		role = user.RoleUser
	}
	// Find user by email
	foundUser, err := s.userService.GetUserByEmail(email)
	
	if err != nil {
		if err.Error() == "user not found" || errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Invalid email or password")
		}
		return err
	}

	// Check if user role matches
	if foundUser.Role != role {
		return errors.New("Invalid email or password")
	}

	// Verify the OTP
	if err := s.otpService.VerifyOtp(email, otpString, otpdto.PurposeResetPwd); err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update the user password in DB
	result := db.DB.Model(&user.User{}).Where("email = ?", email).Update("password", hashedPassword)
	if result.Error != nil {
		return errors.New("failed to update password")
	}

	return nil
}
