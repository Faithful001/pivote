package program

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Program struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	AccessCode  string    `gorm:"type:varchar(100);not null" json:"access_code"`
	IsActive    bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Program) TableName() string {
	return "programs"
}

func (program *Program) BeforeCreate(tx *gorm.DB) error {
	if program.ID == uuid.Nil {
		program.ID = uuid.New()
	}
	if program.AccessCode == "" {
		program.AccessCode = generateRandom4DigitCode()
	}
	return nil
}

func generateRandom4DigitCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "1234"
	}
	return fmt.Sprintf("%04d", n.Int64())
}