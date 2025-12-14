package dtos

import "github.com/google/uuid"

type CreateCandidateDto struct {
	Name      string    `json:"name" binding:"required"`
	ProgramID uuid.UUID `json:"program_id" binding:"required"`
}
