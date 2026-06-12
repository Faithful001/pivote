package dto

type CreateProgramDto struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description" binding:"required"`
	WorkspaceID  string `json:"workspace_id" binding:"required"`
	VotingEndsAt string `json:"voting_ends_at" binding:"required"`
}