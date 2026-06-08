package dto

type UpdateUserDto struct {
	Name  string `json:"name"  binding:"omitempty,min=2,max=100"`
}
