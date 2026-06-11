package sse

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ProgramBroadcaster represents a broadcaster for a single program
type ProgramBroadcaster struct {
	ProgramID    uuid.UUID
	VotingEndsAt time.Time
	IsActive     bool
	Clients      map[chan int64]bool
	StopChan     chan struct{}
}

// BroadcasterManager manages multiple program broadcasters
type BroadcasterManager struct {
	mu           sync.Mutex
	broadcasters map[uuid.UUID]*ProgramBroadcaster
}

// NewBroadcasterManager creates a new BroadcasterManager
func NewBroadcasterManager() *BroadcasterManager {
	return &BroadcasterManager{
		broadcasters: make(map[uuid.UUID]*ProgramBroadcaster),
	}
}

// Register adds a client to a program's broadcaster
func (m *BroadcasterManager) Register(programID uuid.UUID, votingEndsAt time.Time, isActive bool) chan int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.broadcasters[programID]
	if !exists {
		log.Printf("[SSE] Creating broadcaster for program %s (ends at %s, active: %t)", programID, votingEndsAt, isActive)
		b = &ProgramBroadcaster{
			ProgramID:    programID,
			VotingEndsAt: votingEndsAt,
			IsActive:     isActive,
			Clients:      make(map[chan int64]bool),
			StopChan:     make(chan struct{}),
		}
		m.broadcasters[programID] = b

		// Start ticker if program is active and has time remaining
		if isActive && time.Until(votingEndsAt) > 0 {
			go m.runTicker(programID, b.StopChan)
		}
	}

	// Create buffered channel to prevent slow clients from blocking broadcaster
	ch := make(chan int64, 2)
	b.Clients[ch] = true
	log.Printf("[SSE] Client registered to program %s. Total clients: %d", programID, len(b.Clients))
	return ch
}

// Unregister removes a client from a program's broadcaster
func (m *BroadcasterManager) Unregister(programID uuid.UUID, ch chan int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.broadcasters[programID]
	if !exists {
		return
	}

	if _, ok := b.Clients[ch]; ok {
		delete(b.Clients, ch)
		close(ch)
	}

	log.Printf("[SSE] Client unregistered from program %s. Total clients remaining: %d", programID, len(b.Clients))

	// Auto-teardown if no clients are left
	if len(b.Clients) == 0 {
		log.Printf("[SSE] No clients remaining. Tearing down broadcaster for program %s", programID)
		close(b.StopChan)
		delete(m.broadcasters, programID)
	}
}

// UpdateState updates a program's status and end time (called when RabbitMQ state changes are received)
func (m *BroadcasterManager) UpdateState(programID uuid.UUID, votingEndsAt time.Time, isActive bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.broadcasters[programID]
	if !exists {
		// No local listeners for this program, no local broadcaster needed
		return
	}

	b.VotingEndsAt = votingEndsAt
	wasActive := b.IsActive
	b.IsActive = isActive

	log.Printf("[SSE] State update received for program %s: active=%t, endsAt=%s (wasActive=%t)", programID, isActive, votingEndsAt, wasActive)

	// If deactivated or expired, shut down the broadcaster and disconnect clients
	if !isActive || time.Until(votingEndsAt) <= 0 {
		log.Printf("[SSE] Program %s is now inactive or expired. Disconnecting all clients.", programID)
		for ch := range b.Clients {
			select {
			case ch <- 0: // Send 0 to signify end
			default:
			}
			close(ch)
		}
		close(b.StopChan)
		delete(m.broadcasters, programID)
		return
	}

	// If transitioned from inactive to active, spin up the ticker
	if !wasActive && isActive {
		close(b.StopChan) // Stop old ticker if any
		b.StopChan = make(chan struct{})
		go m.runTicker(programID, b.StopChan)
	}
}

// GetBroadcasterCount returns the number of active broadcasters (useful for testing)
func (m *BroadcasterManager) GetBroadcasterCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.broadcasters)
}

// runTicker runs the 1-second countdown loop for a program
func (m *BroadcasterManager) runTicker(programID uuid.UUID, stopChan chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Printf("[SSE] Ticker goroutine started for program %s", programID)
	defer log.Printf("[SSE] Ticker goroutine stopped for program %s", programID)

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			b, exists := m.broadcasters[programID]
			// Ensure we are operating on the correct current broadcaster
			if !exists || b.StopChan != stopChan {
				m.mu.Unlock()
				return
			}

			remaining := int64(time.Until(b.VotingEndsAt).Seconds())

			// Handle end of voting
			if remaining <= 0 || !b.IsActive {
				log.Printf("[SSE] Ticker countdown completed or program deactivated for program %s. Cleaning up.", programID)
				for ch := range b.Clients {
					select {
					case ch <- 0:
					default:
					}
					close(ch)
				}
				close(b.StopChan)
				delete(m.broadcasters, programID)
				m.mu.Unlock()
				return
			}

			// Copy client channels to send tick outside the lock
			clientChs := make([]chan int64, 0, len(b.Clients))
			for ch := range b.Clients {
				clientChs = append(clientChs, ch)
			}
			m.mu.Unlock()

			// Broadcast tick to all clients asynchronously
			for _, ch := range clientChs {
				select {
				case ch <- remaining:
				default:
					// Drop tick if client buffer is full to prevent blocker hanging
				}
			}

		case <-stopChan:
			return
		}
	}
}
