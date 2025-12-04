package candidate

import (
	"pivote/internal/domains/program"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Candidate struct {
    ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name        string          `gorm:"type:varchar(255);not null" json:"name"`
    ProgramID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"program_id"`
    Program     program.Program `gorm:"foreignKey:ProgramID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
    CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Candidate) TableName() string { return "candidates" }

func (candidate *Candidate) BeforeCreate(tx *gorm.DB) error {
    if candidate.ID == uuid.Nil {
        candidate.ID = uuid.New()
    }
    return nil
}