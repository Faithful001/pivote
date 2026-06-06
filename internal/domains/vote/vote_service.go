package vote

import (
	"errors"
	"log"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/program"
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

type ProgramVotesInfo struct {
	ParticipantsCount	int64			`json:"participants_count"`
	TotalVotes          int64            `json:"total_votes"`
	VotesByCandidate    map[string]int64 `json:"votes_by_candidate"`
	UserVoteCandidateID *string          `json:"user_vote_candidate_id"`
}

func (v *VoteService) ToggleVoteCandidate(
	userID uuid.UUID,
	candidateID uuid.UUID,
) (*Vote, error) {
	// 1. Fetch Candidate
	var c candidate.Candidate
	if err := db.DB.Where("id = ?", candidateID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Candidate not found")
		}
		return nil, err
	}

	// 2. Fetch and verify Program status
	var prog program.Program
	if err := db.DB.Where("id = ?", c.ProgramID).First(&prog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Program not found")
		}
		return nil, err
	}

	if !prog.IsActive {
		return nil, errors.New("Voting is closed for this program")
	}

	// 3. Verify user has joined the program
	var joinedCount int64
	db.DB.Model(&program.UserProgram{}).Where("user_id = ? AND program_id = ?", userID, c.ProgramID).Count(&joinedCount)
	if joinedCount == 0 {
		return nil, errors.New("You must join this program before you can vote")
	}

	var vote Vote

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 4. Check if user already voted in this program
		var existingVote Vote
		err := tx.
			Where("user_id = ? AND program_id = ?", userID, c.ProgramID).
			First(&existingVote).Error

		if err == nil {
			// User has already voted
			if existingVote.CandidateID == candidateID {
				// Idempotent: return the existing vote
				vote = existingVote
				return nil
			}
			return errors.New("You have already voted in this program")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 5. User hasn't voted yet, create the vote
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

	v.socketio.BroadcastLeaderboard(msg)
}

func (v *VoteService) GetProgramVotesInfo(programID, userID uuid.UUID) (*ProgramVotesInfo, error) {
	var votes []Vote
	if err := db.DB.Where("program_id = ?", programID).Find(&votes).Error; err != nil {
		return nil, err
	}

	votesByCandidate := make(map[string]int64)
	var totalVotes int64 = 0

	for _, vote := range votes {
		candidateIDStr := vote.CandidateID.String()
		votesByCandidate[candidateIDStr] = votesByCandidate[candidateIDStr] + 1
		totalVotes++
	}

	var userVote Vote
	var userVoteCandidateID *string = nil
	var participantsCount int64 = 0

	err := db.DB.Model(&program.UserProgram{}).Where("user_id = ? AND program_id = ?", userID, programID).Count(&participantsCount).Error
	if err != nil {
		return nil, err
	}

	err = db.DB.Model(&Vote{}).Where("user_id = ? AND program_id = ?", userID, programID).First(&userVote).Error
	if err == nil {
		candidateIDStr := userVote.CandidateID.String()
		userVoteCandidateID = &candidateIDStr
	}

	return &ProgramVotesInfo{
		ParticipantsCount:   participantsCount,
		TotalVotes:          totalVotes,
		VotesByCandidate:    votesByCandidate,
		UserVoteCandidateID: userVoteCandidateID,
	}, nil
}

func (v *VoteService) GetVotesByProgramID(program_id uuid.UUID) ([]Vote, error) {
	var votes []Vote
	result := db.DB.Where("program_id = ?", program_id).Find(&votes)
	if result.Error != nil {
		return nil, result.Error
	}
	return votes, nil
}

func (v *VoteService) GetVotesByCandidateID(candidate_id uuid.UUID) ([]Vote, error) {
	var votes []Vote
	result := db.DB.Where("candidate_id = ?", candidate_id).Find(&votes)
	if result.Error != nil {
		return nil, result.Error
	}
	return votes, nil
}