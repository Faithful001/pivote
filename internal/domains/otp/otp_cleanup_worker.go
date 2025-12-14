package otp

import (
	"log"
	"time"
)

type OtpCleanupWorker struct {
	otpService *OtpService
	ticker     *time.Ticker
	done       chan bool
}

func NewOtpCleanupWorker(otpService *OtpService) *OtpCleanupWorker {
	return &OtpCleanupWorker{
		otpService: otpService,
		ticker:     time.NewTicker(5 * time.Minute),
		done:       make(chan bool),
	}
}

func (w *OtpCleanupWorker) Start() {
	log.Println("OTP cleanup worker started")
	
	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.cleanup()
			case <-w.done:
				log.Println("OTP cleanup worker stopped")
				return
			}
		}
	}()
}

func (w *OtpCleanupWorker) cleanup() {
	count, err := w.otpService.CleanupExpiredOtps()
	if err != nil {
		log.Printf("Error cleaning up expired OTPs: %v", err)
		return
	}
	
	if count > 0 {
		log.Printf("Cleaned up %d expired OTP(s)", count)
	}
}

func (w *OtpCleanupWorker) Stop() {
	w.ticker.Stop()
	w.done <- true
}
