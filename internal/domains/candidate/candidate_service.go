package candidate

import (
	"errors"
	"fmt"
	"pivote/internal/db"
	"pivote/internal/domains/candidate/dtos"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CandidateService struct{}

func NewCandidateService() *CandidateService {
	return &CandidateService{}
}

func (candidate *CandidateService) CreateCandidate(payload dtos.CreateCandidateDto) (*Candidate, error) {
	newCandidate := Candidate{
		Name:      payload.Name,
		ProgramID: payload.ProgramID,
	}

	// Create the candidate in the database
	result := db.DB.Create(&newCandidate)
	if result.Error != nil {
		return nil, result.Error
	}

	return &newCandidate, nil
}

func (candidate *CandidateService) GetCandidates() ([]Candidate, error) {
	var candidates []Candidate
	result := db.DB.Preload("Program").Find(&candidates)
	if result.Error != nil {
		return nil, result.Error
	}
	return candidates, nil
}

func (candidate *CandidateService) GetCandidateById(id uuid.UUID) (*Candidate, error) {
	// Query the database to get candidate by the id provided
	var foundCandidate Candidate
	result := db.DB.Preload("Program").Where("id = ?", id).First(&foundCandidate)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Candidate not found")
	}

	if result.Error != nil {
		return nil, fmt.Errorf("Database error: %v", result.Error)
	}

	return &foundCandidate, nil
}

func (candidate *CandidateService) UpdateCandidate(id uuid.UUID, payload dtos.UpdateCandidateDto) (*Candidate, error) {
	var existingCandidate Candidate

	result := db.DB.Where("id = ?", id).First(&existingCandidate)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Candidate with the provided id not found")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	existingCandidate.Name = payload.Name

	result = db.DB.Save(&existingCandidate)

	if result.Error != nil {
		return nil, result.Error
	}

	return &existingCandidate, nil
}

func (candidate *CandidateService) DeleteCandidate(id uuid.UUID) (*Candidate, error) {
	var existingCandidate Candidate

	result := db.DB.Where("id = ?", id).First(&existingCandidate)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Candidate with the provided id not found")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	result = db.DB.Delete(&existingCandidate)

	if result.Error != nil {
		return nil, result.Error
	}

	return &existingCandidate, nil
}