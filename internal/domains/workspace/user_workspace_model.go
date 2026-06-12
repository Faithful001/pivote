package workspace

import (
	"pivote/internal/domains/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserWorkspace struct {
	ID 			uuid.UUID	`gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID		uuid.UUID	`gorm:"type:uuid;not null" json:"user_id"`
	User		user.User	`gorm:"foreignKey:UserID;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	WorkspaceID	uuid.UUID	`gorm:"type:uuid;not null" json:"workspace_id"`
	Workspace	Workspace	`gorm:"foreignKey:WorkspaceID;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (UserWorkspace) TableName() string {
	return "user_workspaces"
}

func (uw *UserWorkspace) BeforeCreate (tx *gorm.DB) error {
	if uw.ID == uuid.Nil {
		uw.ID = uuid.New()
	} 

	if uw.WorkspaceID == uuid.Nil {
		uw.WorkspaceID = uuid.New()
	} 

	return nil
}