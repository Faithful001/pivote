package user

import (
	"errors"
	"pivote/internal/domains/user/dto"
	"pivote/internal/infra/db"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) CreateUser(user *User) (*User, error) {
	var existingUser User
	result := db.DB.Where("email = ?", user.Email).First(&existingUser)
	
	if result.Error == nil {
		return nil, errors.New("User with this email already exists")
	}
	
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *UserService) GetMe(workspaceID *uuid.UUID, user User) (map[string]any, error) {
	var foundWorkspace WorkspaceInfo

	if workspaceID == nil {
		result := db.DB.Table("workspaces").
			Where("owner_id = ?", user.ID).
			Order("created_at ASC").
			First(&foundWorkspace)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return map[string]any{
					"workspace": nil,
					"user":      user,
				}, nil
			}
			return nil, result.Error
		}
	} else {
		result := db.DB.Table("workspaces").Where("id = ?", *workspaceID).First(&foundWorkspace)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil, errors.New("workspace not found")
			}
			return nil, result.Error
		}
	}

	return map[string]any{
		"workspace": foundWorkspace,
		"user":      user,
	}, nil
}


func (s *UserService) GetUserByID(id uuid.UUID) (*User, error) {
	var user User
	result := db.DB.Where("id = ?", id).First(&user)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	
	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*User, error) {
	var user User
	result := db.DB.Where("email = ?", email).First(&user)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	
	return &user, nil
}

func (s *UserService) GetAllUsers() ([]User, error) {
	var users []User
	result := db.DB.Find(&users)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return users, nil
}

func (s *UserService) UpdateUser(id uuid.UUID, payload dto.UpdateUserDto) (*User, error) {
	var existing User
	if err := db.DB.Where("id = ?", id).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Only update fields the caller is allowed to change
	if payload.Name != "" {
		existing.Name = payload.Name
	}

	if err := db.DB.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	result := db.DB.Where("id = ?", id).Delete(&User{})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	
	return nil
}
