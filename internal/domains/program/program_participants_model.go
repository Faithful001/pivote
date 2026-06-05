package program

import (
	"time"

	"pivote/internal/domains/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProgramParticipant struct {
	ID 			uuid.UUID	`gorm:"type:uuid;primaryKey:default:gen_random_uuid()" json:"id"`
	ProgramID	uuid.UUID	`gorm:"type:uuid;not null" json:"program_id"`
	Program		Program		`gorm:"foreignKey:ProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	UserID		uuid.UUID	`gorm:"type:uuid;not null" json:"user_id"`
	User		user.User	`gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	JoinedAt	time.Time	`gorm:"autoCreateTime" json:"joined_at"`
}

func (ProgramParticipant) TableName () string {
	return "program_participants"
}

func (p *ProgramParticipant) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}