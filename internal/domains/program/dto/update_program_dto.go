package dto

type UpdateProgramDto struct {
	Name         string `json:"name" binding:"omitempty"`
	Description  string `json:"description" binding:"omitempty"`
	VotingEndsAt string `json:"voting_ends_at" binding:"omitempty"`
}