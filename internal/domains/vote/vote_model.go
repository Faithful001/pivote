package vote

import (
	"time"

	"pivote/internal/domains/candidate"
	"pivote/internal/domains/program"
	"pivote/internal/domains/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vote struct {
	ID				uuid.UUID				`gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CandidateID 	uuid.UUID 				`gorm:"type:uuid;not null" json:"candidate_id"`
	Candidate		candidate.Candidate 	`gorm:"foreignKey:CandidateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	UserID 			uuid.UUID 				`gorm:"type:uuid;not null" json:"user_id"`
	User			user.User 				`gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	ProgramID 		uuid.UUID 				`gorm:"type:uuid;not null" json:"program_id"`
	Program			program.Program 		`gorm:"foreignKey:ProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	CreatedAt 		time.Time 				`gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt 		time.Time 				`gorm:"autoUpdateTime" json:"updated_at"`
}

func (Vote) TableName() string {
	return "votes"
}

func (vote *Vote) BeforeCreate(tx *gorm.DB) error {
	if vote.ID == uuid.Nil {
		vote.ID = uuid.New()
	}
	return nil
}