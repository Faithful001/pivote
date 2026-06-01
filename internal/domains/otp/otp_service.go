package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"pivote/internal/domains/otp/dto"
	"pivote/internal/infra/db"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/utils"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OtpService struct {
	mq *rabbitmq.RabbitMQ
}

func NewOtpService(mq *rabbitmq.RabbitMQ) *OtpService {
	return &OtpService{
		mq: mq,
	}
}

func (s *OtpService) GenerateOtp(email string) (string, error) {
	// Generate random number between 0 and 9999
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Convert to integer (0-65535) and modulo 10000
	otpNum := int(b[0])<<8 | int(b[1])
	otpNum = otpNum % 10000

	// Format as 4-digit string with leading zeros
	return fmt.Sprintf("%04d", otpNum), nil
}

func (s *OtpService) SendOtpToEmail(email string, purpose dto.Purpose) error {
	otp, err := s.GenerateOtp(email)
	if err != nil {
		return err
	}

	//save to db
	hashedOtp, err := utils.HashPassword(otp)

	if err != nil {
		return err
	}

	otpToBeSaved := Otp{
		Email:     email,
		Otp:       hashedOtp,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := db.DB.Save(&otpToBeSaved).Error; err != nil {
		return err
	}

	// Publish to RabbitMQ
	body := fmt.Sprintf(`{"email":"%s","otp":"%s", "purpose":"%s"}`, email, otp, purpose)

	err = s.mq.Publish(context.Background(), rabbitmq.PublishConfig{
		Exchange:   "", 
		RoutingKey: "email_otp",
		Mandatory:  false,
		Immediate:  false,
		Message: amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}
	

	return nil
}

func (s *OtpService) VerifyOtp(email string, otp string, purpose dto.Purpose) error {
	var otpFromDb Otp

	result := db.DB.Where("email = ? AND purpose = ?", email, purpose).First(&otpFromDb)

	if result.Error != nil {
		return fmt.Errorf("invalid OTP")
	}

	// Check if OTP has expired
	if time.Now().After(otpFromDb.ExpiresAt) {
		// Delete expired OTP
		db.DB.Delete(&otpFromDb)
		return fmt.Errorf("OTP has expired")
	}

	// Verify the OTP
	if err := utils.VerifyPassword(otpFromDb.Otp, otp); err != nil {
		return fmt.Errorf("invalid OTP")
	}

	// Delete OTP after successful verification to prevent reuse
	if err := db.DB.Delete(&otpFromDb).Error; err != nil {
		return fmt.Errorf("failed to delete OTP: %w", err)
	}

	return nil
}

// CleanupExpiredOtps deletes all expired OTP records from the database
func (s *OtpService) CleanupExpiredOtps() (int64, error) {
	result := db.DB.Where("expires_at < ?", time.Now()).Delete(&Otp{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired OTPs: %w", result.Error)
	}
	return result.RowsAffected, nil
}
