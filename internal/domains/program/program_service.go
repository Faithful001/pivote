package program

import (
	"errors"
	"fmt"
	"pivote/internal/db"
	"pivote/internal/domains/program/dtos"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProgramService struct {}

func NewProgramService() *ProgramService{
	return &ProgramService{}
}

func (program *ProgramService) CreateProgram (payload dtos.CreateProgramDto) (*Program, error) {
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

func (program *ProgramService) GetPrograms() ([]Program, error) {
	var programs []Program
	result := db.DB.Find(&programs)
	if result.Error != nil {
		return nil, result.Error
	}
	return programs, nil
}

func (program *ProgramService) GetProgramById(id uuid.UUID) (*Program, error) {
	//query the database to get program by the id provided
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

func (program *ProgramService) UpdateProgram(id uuid.UUID, payload dtos.UpdateProgramDto) (*Program, error){
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

func (program *ProgramService) DeleteProgram(id uuid.UUID) (*Program, error){
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