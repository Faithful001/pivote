package program

import (
	"context"
	"errors"
	"fmt"
	dtos "pivote/internal/domains/program/dto"
	"pivote/internal/infra/db"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/utils"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type ProgramService struct{}

func NewProgramService() *ProgramService {
	return &ProgramService{}
}

type ProgramResponse struct {
	Program
	IsJoined bool `json:"is_joined"`
}

func (program *ProgramService) CreateProgram(payload dtos.CreateProgramDto) (*Program, error) {
	newProgram := Program{
		Name:        payload.Name,
		Description: payload.Description,
	}

	// Create the program in the database
	result := db.DB.Create(&newProgram)
	if result.Error != nil {
		return nil, result.Error
	}

	return &newProgram, nil
}

func (program *ProgramService) GetPrograms(userID uuid.UUID) ([]ProgramResponse, error) {
	var response []ProgramResponse

	err := db.DB.Model(&Program{}).
		Select("programs.*, CASE WHEN up.program_id IS NOT NULL THEN true ELSE false END AS is_joined").
		Joins("LEFT JOIN user_programs up ON up.program_id = programs.id AND up.user_id = ?", userID).
		Find(&response).Error

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (program *ProgramService) GetProgramById(id uuid.UUID) (*Program, error) {
	// query the database to get program by the id provided
	var foundProgram Program
	result := db.DB.Where("id = ?", id).First(&foundProgram)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Program not found")
	}

	if result.Error != nil {
		return nil, fmt.Errorf("Database error: %v", result.Error)
	}

	return &foundProgram, nil
}

func (program *ProgramService) UpdateProgram(id uuid.UUID, payload dtos.UpdateProgramDto) (*Program, error) {
	var existingProgram Program

	result := db.DB.Where("id = ?", id).First(&existingProgram)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Program with the provided id not found")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	existingProgram.Name = payload.Name
	existingProgram.Description = payload.Description

	result = db.DB.Save(&existingProgram)

	if result.Error != nil {
		return nil, result.Error
	}

	return &existingProgram, nil
}

func (program *ProgramService) DeleteProgram(id uuid.UUID) (*Program, error) {
	var existingProgram Program

	result := db.DB.Where("id = ?", id).First(&existingProgram)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Program with the provided id not found")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	result = db.DB.Delete(&existingProgram)

	if result.Error != nil {
		return nil, result.Error
	}

	return &existingProgram, nil
}

func (program *ProgramService) JoinProgram(userID, programID uuid.UUID, accessCode string) error {
	// 1. Get the program
	var foundProgram Program
	if err := db.DB.Where("id = ?", programID).First(&foundProgram).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Program not found")
		}
		return err
	}

	if !foundProgram.IsActive {
		return errors.New("Program is closed")
	}

	// 2. Validate access code
	var pac ProgramAccessCode
	if err := db.DB.Where("user_id = ? AND program_id = ?", userID, programID).First(&pac).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Access code not requested for this program")
		}
		return err
	}

	decrypted, err := utils.Decrypt(pac.AccessCode, utils.GetEncryptionKey())
	if err != nil {
		return errors.New("Failed to decrypt access code")
	}

	if decrypted != accessCode {
		return errors.New("Invalid access code")
	}

	if pac.IsUsed {
		return errors.New("Access code has already been used")
	}

	// 3. Check if already joined
	var count int64
	db.DB.Model(&UserProgram{}).Where("user_id = ? AND program_id = ?", userID, programID).Count(&count)
	if count > 0 {
		return nil // Idempotent: already joined
	}

	// 4. Mark access code as used
	pac.IsUsed = true
	if err := db.DB.Save(&pac).Error; err != nil {
		return err
	}

	// 5. Create enrollment records in both UserProgram
	enrollment := UserProgram{
		UserID:    userID,
		ProgramID: programID,
	}
	if err := db.DB.Create(&enrollment).Error; err != nil {
		return err
	}

	return nil
}

func (program *ProgramService) RequestVoteCode(userEmail string, userID, programID uuid.UUID, mq *rabbitmq.RabbitMQ) error {
	// 1. Get the program
	var foundProgram Program
	if err := db.DB.Where("id = ?", programID).First(&foundProgram).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Program not found")
		}
		return err
	}

	if !foundProgram.IsActive {
		return errors.New("Program is closed")
	}

	// 2. Check if already joined
	var count int64
	db.DB.Model(&UserProgram{}).Where("user_id = ? AND program_id = ?", userID, programID).Count(&count)
	if count > 0 {
		return errors.New("You have already joined this program")
	}

	// 3. Get or create ProgramAccessCode for this user/program
	var pac ProgramAccessCode
	err := db.DB.Where("user_id = ? AND program_id = ?", userID, programID).First(&pac).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pac = ProgramAccessCode{
				UserID:    userID,
				ProgramID: programID,
			}
			if err := db.DB.Create(&pac).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		pac.IsUsed = false
		code := GenerateRandom4DigitCode()
		encrypted, err := utils.Encrypt(code, utils.GetEncryptionKey())
		if err != nil {
			return err
		}
		pac.AccessCode = encrypted
		if err := db.DB.Save(&pac).Error; err != nil {
			return err
		}
	}

	// 4. Decrypt access code
	decrypted, err := utils.Decrypt(pac.AccessCode, utils.GetEncryptionKey())
	if err != nil {
		return errors.New("Failed to decrypt access code")
	}

	// 5. Publish to RabbitMQ
	body := fmt.Sprintf(`{"email":"%s","otp":"%s","purpose":"%s"}`, userEmail, decrypted, "request_vote")

	err = mq.Publish(context.Background(), rabbitmq.PublishConfig{
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

func (program *ProgramService) ToggleProgram(programID uuid.UUID, isActive bool) (*Program, error) {
	var foundProgram Program
	if err := db.DB.Where("id = ?", programID).First(&foundProgram).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Program not found")
		}
		return nil, err
	}

	foundProgram.IsActive = isActive
	if err := db.DB.Save(&foundProgram).Error; err != nil {
		return nil, err
	}

	return &foundProgram, nil
}