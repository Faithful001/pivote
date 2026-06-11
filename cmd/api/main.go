package main

import (
	"encoding/json"
	"log"
	"os"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/otp"
	"pivote/internal/domains/program"
	"pivote/internal/domains/user"
	"pivote/internal/domains/vote"
	"pivote/internal/infra/db"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/infra/sse"
	"pivote/internal/infra/websocket"
	"pivote/internal/router"
	"pivote/internal/workers/email"
	otpWorker "pivote/internal/workers/otp"
	"time"

	"github.com/google/uuid"
)

func main() {
	// Initialize database connection
	db.InitDB()

	// Initialize rabbitmq
	amqpURL := os.Getenv("RABBITMQ_URL")
	mq, err := rabbitmq.NewRabbitMQ(rabbitmq.Config{
		URL:            amqpURL,
		ReconnectDelay: 5 * time.Second,
		MaxReconnect:   5,
	})
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer mq.Close()

	// Declare email queues
	_, err = mq.DeclareQueue(rabbitmq.QueueConfig{
		Name:    "email.transactional",
		Durable: true,
	})
	if err != nil {
		log.Printf("Failed to declare queue email.transactional: %v", err)
	}

	_, err = mq.DeclareQueue(rabbitmq.QueueConfig{
		Name:    "email.notifications",
		Durable: true,
	})
	if err != nil {
		log.Printf("Failed to declare queue email.notifications: %v", err)
	}

	// Start Email Workers
	transactionalConsumer := email.NewEmailConsumer(mq, "email.transactional")
	transactionalConsumer.Start()

	notificationConsumer := email.NewEmailConsumer(mq, "email.notifications")
	notificationConsumer.Start()

	// Start OTP Cleanup Worker
	otpService := otp.NewOtpService(mq)
	otpCleanupWorker := otpWorker.NewOtpCleanupWorker(otpService)
	otpCleanupWorker.Start()

	// Initialize Socket.IO Server
	ioServer, err := websocket.NewSocketIOServer()
	if err != nil {
		log.Fatalf("Failed to initialize Socket.IO server: %v", err)
	}

	// Initialize SSE Broadcaster Manager
	sseBroadcaster := sse.NewBroadcasterManager()

	// Declare program events fanout exchange
	err = mq.DeclareExchange("program.events", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare exchange program.events: %v", err)
	}

	// Declare a unique auto-deleted exclusive queue for this instance
	q, err := mq.DeclareQueue(rabbitmq.QueueConfig{
		Name:       "", // RabbitMQ generates a unique queue name
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
	})
	if err != nil {
		log.Fatalf("Failed to declare queue for program events: %v", err)
	}

	// Bind the queue to the exchange
	err = mq.BindQueue(q.Name, "", "program.events", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind queue to program.events: %v", err)
	}

	// Start consuming program status change events in a background goroutine
	msgs, err := mq.Consume(rabbitmq.ConsumeConfig{
		Queue:     q.Name,
		Consumer:  "",
		AutoAck:   true,
		Exclusive: true,
	})
	if err != nil {
		log.Printf("Failed to consume program events: %v", err)
	} else {
		go func() {
			log.Println("[Main] Listening for program status change events on RabbitMQ...")
			for d := range msgs {
				var payload struct {
					ProgramID    string `json:"program_id"`
					IsActive     bool   `json:"is_active"`
					VotingEndsAt string `json:"voting_ends_at"`
				}
				if err := json.Unmarshal(d.Body, &payload); err != nil {
					log.Printf("[Main] Error unmarshalling program event: %v", err)
					continue
				}

				pID, err := uuid.Parse(payload.ProgramID)
				if err != nil {
					log.Printf("[Main] Invalid program ID in event: %v", err)
					continue
				}

				var endsAt time.Time
				if payload.VotingEndsAt != "" {
					endsAt, err = time.Parse(time.RFC3339, payload.VotingEndsAt)
					if err != nil {
						log.Printf("[Main] Invalid voting_ends_at format in event: %v", err)
						continue
					}
				}

				sseBroadcaster.UpdateState(pID, endsAt, payload.IsActive)
			}
		}()
	}

	// Run migrations
	if err := db.AutoMigrate(&user.User{}, &program.Program{}, &program.UserProgram{}, &program.ProgramAccessToken{}, &candidate.Candidate{}, &otp.Otp{}, &vote.Vote{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Setup router with SSE Broadcaster Manager
	r := router.SetupRouter(mq, ioServer, sseBroadcaster)

	// Start server
	r.Run(":8000")
}