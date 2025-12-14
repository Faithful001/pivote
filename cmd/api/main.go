package main

import (
	"log"
	"pivote/internal/db"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/otp"
	"pivote/internal/domains/program"
	"pivote/internal/domains/user"
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
	mq, err := rabbitmq.NewRabbitMQ(rabbitmq.Config{
		URL:            "amqp://guest:guest@localhost:5672/",
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

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Run migrations
	if err := db.AutoMigrate(&user.User{}, &program.Program{}, &candidate.Candidate{}, &otp.Otp{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Setup router
	r := router.SetupRouter(mq, hub)

	// Start server
	r.Run(":8000")
}