package workspace

import (
	"errors"
	"pivote/internal/domains/workspace/dto"
	"pivote/internal/infra/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceService struct{}

func NewWorkspaceService() WorkspaceService {
	return WorkspaceService{}
}

// admin only
func (ws *WorkspaceService) CreateWorkspace(userID uuid.UUID, payload dto.CreateWorkspaceDto) (*Workspace, error) {
	newWorkspace := Workspace {
		Name: payload.Name,
		OwnerID: userID,
	}

	result := db.DB.Create(&newWorkspace)
	
	if result.Error != nil {
		return nil, result.Error
	}

	newUserWorkspace := UserWorkspace {
		UserID: userID,
		WorkspaceID: newWorkspace.ID,
	}

	if err := db.DB.Create(&newUserWorkspace).Error; err != nil {
		return nil, err
	}

	return &newWorkspace, nil
}


// admin only
func (ws *WorkspaceService) GetWorkspaces(userID uuid.UUID) (*[]Workspace, error) {
	var workspaces []Workspace

	if err := db.DB.Where("owner_id = ?", userID).Find(&workspaces).Error; err != nil {
		return nil, err
	}

	return &workspaces, nil
}

// admin only
func (ws *WorkspaceService) GetWorkspace(userID uuid.UUID, workspaceID uuid.UUID) (*Workspace, error) {
	var workspace Workspace

	err := db.DB.Where("id = ? AND owner_id = ?", workspaceID, userID).First(&workspace).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("Workspace does not exist")
	} 

	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

// admin only
func (ws *WorkspaceService) UpdateWorkspace(userID uuid.UUID, workspaceID uuid.UUID, payload dto.UpdateWorkspaceDto) error {
	var workspace Workspace

	err := db.DB.Where("id = ? AND owner_id = ?", workspaceID, userID).First(&workspace).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("Workspace does not exist")
	} 

	if err != nil {
		return err
	}

	workspace.Name = payload.Name

	if err := db.DB.Save(&workspace).Error; err != nil {
		return err
	}

	return nil
}

// admin only
func (ws *WorkspaceService) DeleteWorkspace(userID uuid.UUID, workspaceID uuid.UUID) error {
	var workspace Workspace

	result := db.DB.Where("id = ? AND owner_id = ?", workspaceID, userID).First(&workspace)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("Workspace does not exist")
	} 

	if result.Error != nil {
		return result.Error
	}

	if err := db.DB.Delete(&workspace).Error; err != nil {
		return err
	}

	return nil
}
