package main

import (
	"log"
	"os"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/otp"
	"pivote/internal/domains/program"
	"pivote/internal/domains/user"
	"pivote/internal/domains/vote"
	"pivote/internal/infra/db"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/infra/websocket"
	"pivote/internal/router"
	"pivote/internal/workers/email"
	otpWorker "pivote/internal/workers/otp"
	"time"
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

		// Run migrations
	if err := db.AutoMigrate(&user.User{}, &program.Program{}, &program.UserProgram{}, &program.ProgramAccessToken{}, &candidate.Candidate{}, &otp.Otp{}, &vote.Vote{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Setup router
	r := router.SetupRouter(mq, ioServer)

	// Start server
	r.Run(":8000")
}