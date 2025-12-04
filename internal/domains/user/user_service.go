package user

import (
	"errors"
	"pivote/internal/db"

	"gorm.io/gorm"
)

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

func (s *UserService) GetUserByID(id uint) (*User, error) {
	var user User
	result := db.DB.First(&user, id)
	
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

func (s *UserService) UpdateUser(user *User) error {
	result := db.DB.Save(user)
	return result.Error
}

func (s *UserService) DeleteUser(id uint) error {
	result := db.DB.Delete(&User{}, id)
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	
	return nil
}
