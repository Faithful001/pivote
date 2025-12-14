package vote

import (
	"pivote/internal/db"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteService struct{}

func NewVoteService() *VoteService {
	return &VoteService{}
}

func (v *VoteService) ToggleVoteCandidate(user_id uuid.UUID, candidate_id uuid.UUID) (*Vote, error) {
	// Check if a vote exists with the user_id and candidate_id
	var vote Vote

	result := db.DB.Where("user_id = ? AND candidate_id = ?", user_id, candidate_id).First(&vote)

	// Vote found -> Delete it (unvote)
	if result.Error == nil {
		deleteResult := db.DB.Delete(&vote)
		if deleteResult.Error != nil {
			return nil, deleteResult.Error
		}
		return &vote, nil
	}

	// Error is NOT "record not found" -> Real database error
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// Vote not found -> Create it (vote)
	newVote := Vote{
		CandidateID: candidate_id,
		UserID:      user_id,
	}

	createdVote := db.DB.Create(&newVote)
	if createdVote.Error != nil {
		return nil, createdVote.Error
	}

	return &newVote, nil
}

func (v *VoteService) GetVotesByProgramID (program_id uuid.UUID) ([]Vote, error){
	var votes []Vote
	
	// select votes where program_id = program_id
	result := db.DB.Where("program_id = ?", program_id).Find(&votes)

	if result.Error != nil {
		return nil, result.Error
	}

	return votes, nil
}

func (v *VoteService) GetVotesByCandidateID (candidate_id uuid.UUID) ([]Vote, error){
	var votes []Vote

	//select votes where candidate_id = candidate_id
	result := db.DB.Where("candidate_id = ?", candidate_id).Find(&votes)

	if result.Error != nil {
		return nil, result.Error
	}

	return votes, nil
}