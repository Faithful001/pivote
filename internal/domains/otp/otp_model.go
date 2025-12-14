package otp

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Purpose string

const (
	PurposeVerifyAcct 	Purpose = "verify_acct"
	PurposeResetPwd 	Purpose = "reset_pwd"
)

type Otp struct {
	ID			uuid.UUID	`gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email		string		`gorm:"type:varchar(100)" json:"email"`
	Otp			string		`gorm:"type:varchar(100)" json:"otp"`
	Purpose 	Purpose		`gorm:"type:varchar(100)" json:"purpose"` 
	ExpiresAt	time.Time	`gorm:"type:timestamp;index" json:"expires_at"`
	CreatedAt	time.Time	`gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt	time.Time	`gorm:"autoUpdateTime" json:"updated_at"`
}

func (Otp) TableName() string {
	return "otps"
}

func (otp *Otp) BeforeCreate(tx *gorm.DB) (err error) {
	if otp.ID == uuid.Nil {
		otp.ID = uuid.New()
	}
	return nil
}