package program

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"pivote/internal/domains/otp/dto"
	programDto "pivote/internal/domains/program/dto"
	"pivote/internal/domains/user"
	"pivote/internal/domains/workspace"
	"pivote/internal/infra/db"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/types"
	"pivote/internal/utils"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)


type ProgramService struct {
	jwtUtil *utils.JWTUtil
	mq      *rabbitmq.RabbitMQ
}

func NewProgramService(mq *rabbitmq.RabbitMQ) (*ProgramService, error) {
	jwtUtil, err := utils.NewJWTUtil(os.Getenv("JWT_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize jwt util: %w", err)
	}

	return &ProgramService{
		jwtUtil: jwtUtil,
		mq:      mq,
	}, nil
}

type ProgramResponse struct {
	Program
	IsJoined bool `json:"is_joined"`
}

// Create Program - admins only
func (program *ProgramService) CreateProgram(payload programDto.CreateProgramDto) (*Program, error) {
	
	//parse the workspaceId to a uuid
	workspaceID, err := uuid.Parse(payload.WorkspaceID)

	if err != nil {
		return nil, errors.New("Invalid workspace_id provided")
	}

	votingEndsAt, err := time.Parse(time.RFC3339, payload.VotingEndsAt)
	if err != nil {
		return nil, errors.New("invalid voting_ends_at format, expected ISO 8601")
	}

	newProgram := Program{
		Name:        payload.Name,
		Description: payload.Description,
		WorkspaceID: workspaceID,
		VotingEndsAt: &votingEndsAt,
	}

	result := db.DB.Create(&newProgram)
	if result.Error != nil {
		return nil, result.Error
	}

	return &newProgram, nil
}

func (program *ProgramService) GetPrograms(userID uuid.UUID, workspaceID uuid.UUID, role user.Role) ([]ProgramResponse, error) {
	var response []ProgramResponse

	var err error
	
	switch role {
	case user.RoleUser:
		err = db.DB.Model(&Program{}).
		Select("programs.*, CASE WHEN up.program_id IS NOT NULL THEN true ELSE false END AS is_joined", workspaceID).
		Where("programs.workspace_id = ?", workspaceID).
		Joins("JOIN user_programs up ON up.program_id = programs.id AND up.user_id = ?", userID).
		Find(&response).Error
	case user.RoleAdmin:
		err = db.DB.Model(&Program{}).
		Select("programs.*, CASE WHEN up.program_id IS NOT NULL THEN true ELSE false END AS is_joined", userID, workspaceID).
		Where("programs.owner_id = ? AND programs.workspace_id = ?", userID, workspaceID).
		Joins("LEFT JOIN user_programs up ON up.program_id = programs.id AND up.user_id = ?", userID).
		Find(&response).Error
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}


func (p *ProgramService) GetProgramById(id uuid.UUID, userID uuid.UUID, role user.Role) (*ProgramResponse, error) {
	var foundProgram ProgramResponse

	query := db.DB.Model(&Program{}).
		Select("programs.*, CASE WHEN up.program_id IS NOT NULL THEN true ELSE false END AS is_joined").
		Where("programs.id = ?", id)

	switch role {
	case user.RoleUser:
		query = query.Joins("JOIN user_programs up ON up.program_id = programs.id AND up.user_id = ?", userID)
	case user.RoleAdmin:
		query = query.Joins("LEFT JOIN user_programs up ON up.program_id = programs.id AND up.user_id = ?", userID)
	}

	if err := query.First(&foundProgram).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("program not found")
		}
		return nil, err
	}

	return &foundProgram, nil
}

func (program *ProgramService) UpdateProgram(id uuid.UUID, payload programDto.UpdateProgramDto) (*Program, error) {
	var existingProgram Program

	result := db.DB.Where("id = ?", id).First(&existingProgram)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Program with the provided id not found")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	existingProgram.Name = payload.Name
	existingProgram.Description = payload.Description

	result = db.DB.Save(&existingProgram)

	if result.Error != nil {
		return nil, result.Error
	}

	return &existingProgram, nil
}

func (program *ProgramService) DeleteProgram(id uuid.UUID) (*Program, error) {
	var existingProgram Program

	result := db.DB.Where("id = ?", id).First(&existingProgram)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Program with the provided id not found")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	result = db.DB.Delete(&existingProgram)

	if result.Error != nil {
		return nil, result.Error
	}

	return &existingProgram, nil
}

func (p *ProgramService) JoinProgram(programID uuid.UUID, workspaceID uuid.UUID, tokenStr string) error {
	claims, err := p.jwtUtil.ParseToken(tokenStr)
	if err != nil {
		return err
	}

	if claims.Purpose != types.JwtPurposeProgramJoin {
		return errors.New("invalid or expired token")
	}

	if claims.ProgramID == nil || *claims.ProgramID != programID {
		return errors.New("invalid or expired token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var foundUser user.User

		if err := tx.Where("id = ?", userID).First(&foundUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid or expired token")
			}
			return err
		}

		var foundProgram Program
		if programErr := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", programID).
			First(&foundProgram).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("program not found")
			}
			return programErr
		}

		var foundWorkspace workspace.Workspace
		if workspaceErr := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", workspaceID).
			First(&foundWorkspace).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("workspace not found")
			}
			return workspaceErr
		}
		// if !foundProgram.IsActive {
		// 	return errors.New("program is closed")
		// }

		var pac ProgramAccessToken
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ? AND program_id = ?", foundUser.ID, programID).
			First(&pac).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid or expired join link")
			}
			return err
		}

		if pac.IsUsed {
			return errors.New("invalid or expired join link")
		}

		var count int64
		if err := tx.Model(&UserProgram{}).
			Where("user_id = ? AND program_id = ?", foundUser.ID, programID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		pac.IsUsed = true
		if err := tx.Save(&pac).Error; err != nil {
			return err
		}



		// add the user to the corresponding workspace
		if err := tx.Create(&workspace.UserWorkspace {
			UserID: foundUser.ID,
			WorkspaceID: workspaceID,
		}).Error; err != nil {
			return err
		} 

		return tx.Create(&UserProgram{
			UserID:    foundUser.ID,
			ProgramID: programID,
		}).Error
	})
}

func (p *ProgramService) RequestJoinLink(userEmail string, programID uuid.UUID, workspaceID uuid.UUID) error {
	// 1. Verify program exists and is active first
	// (no point doing anything if the program is invalid)
	var foundProgram Program
	if err := db.DB.Where("id = ?", programID).First(&foundProgram).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("program not found")
		}
		return err
	}
	// if !foundProgram.IsActive {
	// 	return errors.New("program is closed")
	// }

	var foundWorkspace workspace.Workspace

	if workspaceErr := db.DB.Where("id = ?", workspaceID).First(&foundWorkspace).Error; workspaceErr != nil {
		if errors.Is(workspaceErr, gorm.ErrRecordNotFound){
			return errors.New("Workspace not found")
		}

		return workspaceErr
	}

	// 2. Check if user exists
	var foundUser user.User
	userErr := db.DB.Where("email = ?", userEmail).First(&foundUser).Error

	if userErr != nil && !errors.Is(userErr, gorm.ErrRecordNotFound) {
		return userErr // real DB error
	}

	// user does not exist - send registration nudge email and return
	if errors.Is(userErr, gorm.ErrRecordNotFound) {
		return p.sendRegistrationNudge(userEmail, programID, foundProgram.Name)
	}

	// 3. Check if already enrolled
	var count int64
	if err := db.DB.Model(&UserProgram{}).
		Where("user_id = ? AND program_id = ?", foundUser.ID, programID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("you have already joined this program")
	}

	// 4. Generate a fresh JWT
	expiresAt := time.Now().Add(5 * time.Minute)
	token, err := p.jwtUtil.GenerateToken(utils.TokenOptions{
		UserID:    foundUser.ID,
		Role:      string(user.RoleUser),
		Purpose:   types.JwtPurposeProgramJoin,
		ProgramID: &programID,
		WorkspaceID: &workspaceID,
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		return fmt.Errorf("failed to generate join token: %w", err)
	}

	// 5. Upsert the ProgramAccessToken
	var pac ProgramAccessToken
	err = db.DB.Where("user_id = ? AND program_id = ?", foundUser.ID, programID).First(&pac).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		pac = ProgramAccessToken{
			UserID:      foundUser.ID,
			ProgramID:   programID,
			AccessToken: token,
		}
		if err := db.DB.Create(&pac).Error; err != nil {
			return err
		}
	} else {
		pac.AccessToken = token
		pac.IsUsed = false
		if err := db.DB.Save(&pac).Error; err != nil {
			return err
		}
	}

	// 6. Build join link
	baseURL := "http://localhost:5173"
	if os.Getenv("ENV") == "production" {
		baseURL = "https://pivote.ng"
	}
	joinLink := fmt.Sprintf("%s/programs/%s/join?token=%s&email=%s&workspace_name=%sprogram_name=%s",
		baseURL,
		programID.String(),
		token,
		url.QueryEscape(userEmail),
		url.QueryEscape(foundWorkspace.Name),
		url.QueryEscape(foundProgram.Name),
	)

	// 7. Publish join link email
	return p.publishEmail(map[string]string{
		"email":     userEmail,
		"join_link": joinLink,
		"purpose":   string(dto.PurposeRequestJoinLink),
	})
}

// sendRegistrationNudge sends an email to unregistered users
// with a link to register before joining the program
func (p *ProgramService) sendRegistrationNudge(userEmail string, programID uuid.UUID, programName string) error {
	baseURL := "http://localhost:5173"
	if os.Getenv("ENV") == "production" {
		baseURL = "https://pivote.ng"
	}

	registerLink := fmt.Sprintf("%s/register?email=%s&program_id=%s&program_name=%s",
		baseURL,
		url.QueryEscape(userEmail),
		programID.String(),
		url.QueryEscape(programName),
	)

	return p.publishEmail(map[string]string{
		"email":         userEmail,
		"register_link": registerLink,
		"program_name":  programName,
		"purpose":       string(dto.PurposeRegisterToJoin),
	})
}

// publishEmail is a helper to avoid repeating the publish block
func (p *ProgramService) publishEmail(msg map[string]string) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal email message: %w", err)
	}

	if err := p.mq.Publish(context.Background(), rabbitmq.PublishConfig{
		Exchange:   "",
		RoutingKey: "email.notifications",
		Mandatory:  false,
		Immediate:  false,
		Message: amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	}); err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}

	return nil
}

func (program *ProgramService) ToggleProgram(programID uuid.UUID, isActive bool, votingEndsAt string) (*Program, error) {
	var foundProgram Program
	err := db.DB.Transaction(func (tx *gorm.DB) error {
		
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", programID).
			First(&foundProgram).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("program not found")
			}
			return err
		}
		
		foundProgram.IsActive = isActive

		if isActive {
			if votingEndsAt == "" {
				return errors.New("voting_ends_at is required when activating a program")
			}
			parsedVotingEndsAt, err := time.Parse(time.RFC3339, votingEndsAt)
			if err != nil {
				return errors.New("invalid voting_ends_at format, expected ISO 8601")
			}
			if parsedVotingEndsAt.Before(time.Now()) {
				return errors.New("voting_ends_at must be in the future")
			}
			foundProgram.VotingEndsAt = &parsedVotingEndsAt
		} else {
			foundProgram.VotingEndsAt = nil
		}
			
		if err := tx.Model(&foundProgram).Updates(map[string]interface{}{
			"is_active": foundProgram.IsActive,
			"voting_ends_at": foundProgram.VotingEndsAt,
		}).Error; err != nil {
			return err
		}
		
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish state change event to RabbitMQ fanout exchange
	if publishErr := program.publishProgramStatusEvent(foundProgram.ID, foundProgram.IsActive, foundProgram.VotingEndsAt); publishErr != nil {
		log.Printf("[ProgramService] Failed to publish program status event: %v", publishErr)
	}

	return &foundProgram, nil
}

func (p *ProgramService) publishProgramStatusEvent(programID uuid.UUID, isActive bool, votingEndsAt *time.Time) error {
	event := map[string]interface{}{
		"program_id":     programID.String(),
		"is_active":      isActive,
		"voting_ends_at": nil,
	}
	if votingEndsAt != nil {
		event["voting_ends_at"] = votingEndsAt.Format(time.RFC3339)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal program status event: %w", err)
	}

	return p.mq.Publish(context.Background(), rabbitmq.PublishConfig{
		Exchange:   "program.events",
		RoutingKey: "",
		Mandatory:  false,
		Immediate:  false,
		Message: amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	})
}