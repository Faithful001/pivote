package dto

type CreateWorkspaceDto struct {
	Name string `json:"name" binding:"required"`
}