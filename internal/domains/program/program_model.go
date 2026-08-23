package program

import (
	"pivote/internal/domains/user"
	"pivote/internal/domains/workspace"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Program struct {
	ID           	uuid.UUID 			`gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         	string    			`gorm:"type:varchar(255);not null" json:"name"`
	Description  	string    			`gorm:"type:text" json:"description"`
	WorkspaceID		uuid.UUID			`gorm:"type:uuid;not null" json:"workspace_id"`
	Workspace		workspace.Workspace	`gorm:"foreignKey:WorkspaceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"workspace"`
	OwnerID			uuid.UUID			`gorm:"type:uuid;not null" json:"owner_id"`
	Owner			user.User			`gorm:"foreignKey:OwnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	IsActive     	bool      			`gorm:"type:boolean;not null;default:false" json:"is_active"`
	VotingEndsAt 	*time.Time 			`gorm:"type:timestamp;nullable" json:"voting_ends_at"`
	CreatedAt    	time.Time 			`gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    	time.Time 			`gorm:"autoUpdateTime" json:"updated_at"`

	// UserPrograms []UserProgram `gorm:"foreignKey:ProgramID" json:"user_programs,omitempty"`
}

func (Program) TableName() string {
	return "programs"
}

func (program *Program) BeforeCreate(tx *gorm.DB) error {
	if program.ID == uuid.Nil {
		program.ID = uuid.New()
	}
	return nil
}