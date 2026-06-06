package program

import (
	"pivote/internal/domains/user"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProgram struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      user.User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	ProgramID uuid.UUID `gorm:"type:uuid;not null" json:"program_id"`
	Program   Program   `gorm:"foreignKey:ProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserProgram) TableName() string {
	return "user_programs"
}

func (u *UserProgram) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
