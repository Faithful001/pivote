package dto

import "github.com/google/uuid"

// LeaderboardEntry represents a single candidate's standing in the leaderboard.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	CandidateID uuid.UUID `json:"candidate_id"`
	Name        string    `json:"name"`
	VoteCount   int64     `json:"vote_count"`
}

// LeaderboardUpdate is the WebSocket message broadcast after every vote toggle.
type LeaderboardUpdate struct {
	Type      string              `json:"type"`       // always "leaderboard_update"
	ProgramID uuid.UUID           `json:"program_id"`
	Data      []LeaderboardEntry  `json:"data"`
}
