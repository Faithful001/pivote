package dtos

import "github.com/google/uuid"

type ToggleVoteDto struct {
	CandidateID uuid.UUID `json:"candidate_id" binding:"required"`
}
