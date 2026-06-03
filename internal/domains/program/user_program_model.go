package program

import (
	"time"

	"github.com/google/uuid"
)

type UserProgram struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	ProgramID uuid.UUID `gorm:"type:uuid;primaryKey" json:"program_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserProgram) TableName() string {
	return "user_programs"
}
