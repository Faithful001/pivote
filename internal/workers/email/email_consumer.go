package email

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"pivote/internal/infra/rabbitmq"

	"pivote/internal/domains/otp/dto"

	ampq "github.com/rabbitmq/amqp091-go"
	"github.com/resend/resend-go/v3"
)

type EmailMessage struct {
	Email 	string `json:"email"`
	Otp   	string `json:"otp"`
	Purpose dto.Purpose `json:"purpose"`
}

type EmailConsumer struct {
	mq     *rabbitmq.RabbitMQ
	client *resend.Client
}

func NewEmailConsumer(mq *rabbitmq.RabbitMQ) *EmailConsumer {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: RESEND_API_KEY is not set. Emails will not be sent.")
	}
	
	return &EmailConsumer{
		mq:     mq,
		client: resend.NewClient(apiKey),
	}
}

func (c *EmailConsumer) Start() {
	msgs, err := c.mq.Consume(rabbitmq.ConsumeConfig{
		Queue:     "email_otp",
		Consumer:  "email_worker",
		AutoAck:   false, // ack manually after sending
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	if err != nil {
		log.Printf("Failed to start email consumer: %v", err)
		return
	}

	go c.processMessages(msgs)
}

func (c *EmailConsumer) processMessages(msgs <-chan ampq.Delivery) {
	log.Println("Email consumer started listening...")
	for d := range msgs {
		var msg EmailMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			d.Ack(false) // Ack anyway to remove bad message
			continue
		}

		err := c.sendEmail(msg)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", msg.Email, err)
			// d.Nack(false, true) // Retry? Or just log and ack?
			// For now, ack to prevent infinite loops if it's a permanent error
			d.Ack(false) 
		} else {
			log.Printf("OTP email sent to %s", msg.Email)
			d.Ack(false)
		}
	}
}

func (c *EmailConsumer) sendEmail(msg EmailMessage) error {
	purposeMap := map[dto.Purpose]string{
		dto.PurposeVerifyAcct: "Verify your Account",
		dto.PurposeResetPwd:   "Reset your Password",
	}

	friendlyPurpose, ok := purposeMap[msg.Purpose]
	if !ok {
		friendlyPurpose = "Authentication"
	}

	params := &resend.SendEmailRequest{
		From:    "Pivote <admin@resend.dev>",
		To:      []string{msg.Email},
		Subject: fmt.Sprintf("%s - Pivote", friendlyPurpose),
		Html:    fmt.Sprintf("<p>Your OTP code to <strong>%s</strong> is: <strong>%s</strong></p>", friendlyPurpose, msg.Otp),
	}

	_, err := c.client.Emails.Send(params)
	if err != nil {
		return err
	}
	return nil
}
