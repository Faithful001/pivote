package program

import (
	"time"

	"pivote/internal/domains/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProgramAccessToken struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User        user.User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	ProgramID   uuid.UUID `gorm:"type:uuid;not null" json:"program_id"`
	Program     Program   `gorm:"foreignKey:ProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	AccessToken  string    `gorm:"type:text;not null" json:"access_token"`
	IsUsed		bool	  `gorm:"type:boolean;not null;default:false" json:"is_used"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProgramAccessToken) TableName() string {
	return "program_access_tokens"
}

func (p *ProgramAccessToken) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	
	return nil
}


