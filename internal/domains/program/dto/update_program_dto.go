package dtos

type UpdateProgramDto struct {
	Name        string `gorm:"type:string" json:"name"`
	Description string `gorm:"type:string" json:"description"`
}