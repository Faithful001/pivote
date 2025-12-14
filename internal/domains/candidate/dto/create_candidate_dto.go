package dto

type CreateCandidateDto struct {
	Name      string `json:"name" binding:"required"`
	ProgramID string `json:"program_id" binding:"required"`
}
