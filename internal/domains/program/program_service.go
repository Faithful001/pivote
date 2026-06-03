package program

import (
	"errors"
	"fmt"
	dtos "pivote/internal/domains/program/dto"
	"pivote/internal/infra/db"

	"github.com/google/uuid"
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
	var programs []Program
	result := db.DB.Find(&programs)
	if result.Error != nil {
		return nil, result.Error
	}

	// Fetch joined program IDs for this user
	var joined []uuid.UUID
	db.DB.Model(&UserProgram{}).Where("user_id = ?", userID).Pluck("program_id", &joined)

	joinedMap := make(map[uuid.UUID]bool)
	for _, pid := range joined {
		joinedMap[pid] = true
	}

	responses := make([]ProgramResponse, len(programs))
	for i, p := range programs {
		responses[i] = ProgramResponse{
			Program:  p,
			IsJoined: joinedMap[p.ID],
		}
	}

	return responses, nil
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

	// 2. Validate access code
	if foundProgram.AccessCode != accessCode {
		return errors.New("Invalid access code")
	}

	// 3. Check if already joined
	var count int64
	db.DB.Model(&UserProgram{}).Where("user_id = ? AND program_id = ?", userID, programID).Count(&count)
	if count > 0 {
		return nil // Idempotent: already joined
	}

	// 4. Create enrollment record
	enrollment := UserProgram{
		UserID:    userID,
		ProgramID: programID,
	}
	if err := db.DB.Create(&enrollment).Error; err != nil {
		return err
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