package dto

type UpdateWorkspaceDto struct {
	Name string `json:"name" binding:"required"`
}