package vote

import (
	"errors"
	"log"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/vote/dtos"
	"pivote/internal/infra/db"
	"pivote/internal/infra/websocket"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteService struct {
	socketio *websocket.SocketIOServer
}

func NewVoteService(socketio *websocket.SocketIOServer) *VoteService {
	return &VoteService{socketio: socketio}
}

func (v *VoteService) ToggleVoteCandidate(
	userID uuid.UUID,
	candidateID uuid.UUID,
) (*Vote, error) {

	var c candidate.Candidate
	if err := db.DB.Where("id = ?", candidateID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("candidate not found")
		}
		return nil, err
	}

	var vote Vote

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.
			Where("user_id = ? AND candidate_id = ?", userID, candidateID).
			First(&vote).Error

		if err == nil {
			// Vote exists — remove it
			return tx.Delete(&vote).Error
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Vote doesn't exist → insert
		vote = Vote{
			UserID:      userID,
			CandidateID: candidateID,
			ProgramID:   c.ProgramID,
		}

		return tx.Create(&vote).Error
	})

	if err != nil {
		return nil, err
	}

	go v.broadcastLeaderboard(c.ProgramID)

	return &vote, nil
}

func (v *VoteService) GetLeaderboard(programID uuid.UUID) ([]dtos.LeaderboardEntry, error) {
	type row struct {
		CandidateID uuid.UUID `gorm:"column:candidate_id"`
		Name        string    `gorm:"column:name"`
		VoteCount   int64     `gorm:"column:vote_count"`
	}

	var rows []row

	err := db.DB.
		Table("votes").
		Select("votes.candidate_id, candidates.name, COUNT(votes.id) AS vote_count").
		Joins("JOIN candidates ON candidates.id = votes.candidate_id").
		Where("votes.program_id = ?", programID).
		Group("votes.candidate_id, candidates.name").
		Order("vote_count DESC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	entries := make([]dtos.LeaderboardEntry, len(rows))
	for i, r := range rows {
		entries[i] = dtos.LeaderboardEntry{
			Rank:        i + 1,
			CandidateID: r.CandidateID,
			Name:        r.Name,
			VoteCount:   r.VoteCount,
		}
	}

	return entries, nil
}

func (v *VoteService) broadcastLeaderboard(programID uuid.UUID) {
	if v.socketio == nil {
		return
	}

	entries, err := v.GetLeaderboard(programID)
	if err != nil {
		log.Printf("leaderboard broadcast: failed to fetch leaderboard for program %s: %v", programID, err)
		return
	}

	msg := dtos.LeaderboardUpdate{
		Type:      "leaderboard:update",
		ProgramID: programID,
		Data:      entries,
	}

	// payload, err := json.Marshal(msg)
	// if err != nil {
	// 	log.Printf("leaderboard broadcast: failed to marshal message: %v", err)
	// 	return
	// }

	v.socketio.BroadcastLeaderboard(msg)
}

func (v *VoteService) GetVotesByProgramID(program_id uuid.UUID) ([]Vote, error) {
	var votes []Vote

	// select votes where program_id = program_id
	result := db.DB.Where("program_id = ?", program_id).Find(&votes)

	if result.Error != nil {
		return nil, result.Error
	}

	return votes, nil
}

func (v *VoteService) GetVotesByCandidateID(candidate_id uuid.UUID) ([]Vote, error) {
	var votes []Vote

	//select votes where candidate_id = candidate_id
	result := db.DB.Where("candidate_id = ?", candidate_id).Find(&votes)

	if result.Error != nil {
		return nil, result.Error
	}

	return votes, nil
}