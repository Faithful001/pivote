package dtos

type CreateProgramDto struct {
	Name        string `gorm:"type:string" json:"name" binding:"required"`
	Description string `gorm:"type:string" json:"description" binding:"required"`
}