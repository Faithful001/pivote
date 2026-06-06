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

	// Declare email queue
	_, err = mq.DeclareQueue(rabbitmq.QueueConfig{
		Name:    "email_otp",
		Durable: true,
	})
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
	}

	// Start Email Worker
	emailConsumer := email.NewEmailConsumer(mq)
	emailConsumer.Start()

	// Start OTP Cleanup Worker
	otpService := otp.NewOtpService(mq)
	otpCleanupWorker := otp.NewOtpCleanupWorker(otpService)
	otpCleanupWorker.Start()

	// Initialize Socket.IO Server
	ioServer, err := websocket.NewSocketIOServer()
	if err != nil {
		log.Fatalf("Failed to initialize Socket.IO server: %v", err)
	}

		// Run migrations
	if err := db.AutoMigrate(&user.User{}, &program.Program{}, &program.UserProgram{}, &program.ProgramAccessCode{}, &candidate.Candidate{}, &otp.Otp{}, &vote.Vote{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Setup router
	r := router.SetupRouter(mq, ioServer)

	// Start server
	r.Run(":8000")
}