package workspace

import (
	"pivote/internal/domains/user"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Workspace struct {
	ID			uuid.UUID	`gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name		string 		`gorm:"type:varchar(255);not null" json:"name"`
	OwnerID		uuid.UUID 		`gorm:"type:uuid;not null" json:"owner_id"`
	Owner		user.User	`gorm:"foreignKey:OwnerID;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCASE" json:"-"`
	CreatedAt	time.Time	`gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt	time.Time	`gorm:"autoUpdateTime" json:"updated_at"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

func (w *Workspace) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}

	return nil
}