package program

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Program struct {
	ID uuid.UUID 		`gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name string 		`gorm:"type:varchar(255);not null" json:"name"`
	Description string 	`gorm:"type:text" json:"description"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}


func (Program) TableName() string {
	return "programs"
}


func (program *Program) BeforeCreate(tx *gorm.DB) error {
	if program.ID == uuid.Nil {
		program.ID = uuid.New()
	}
	return nil
}