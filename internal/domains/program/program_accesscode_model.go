package program

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"pivote/internal/domains/user"
	"pivote/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProgramAccessCode struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User        user.User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	ProgramID   uuid.UUID `gorm:"type:uuid;not null" json:"program_id"`
	Program     Program   `gorm:"foreignKey:ProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	AccessCode  string    `gorm:"type:varchar(100);not null" json:"access_code"`
	IsUsed		bool	  `gorm:"type:boolean;not null;default:false" json:"is_used"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProgramAccessCode) TableName() string {
	return "program_access_codes"
}

func (p *ProgramAccessCode) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	code := GenerateRandom4DigitCode()
	encrypted, err := utils.Encrypt(code, utils.GetEncryptionKey())
	if err != nil {
		return err
	}
	p.AccessCode = encrypted
	return nil
}

func GenerateRandom4DigitCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "1234"
	}
	return fmt.Sprintf("%04d", n.Int64())
}
