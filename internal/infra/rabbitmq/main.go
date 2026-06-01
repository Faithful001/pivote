package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn 			*amqp.Connection
	channel 		*amqp.Channel
	isConnected 	bool
	mu 				sync.RWMutex
	config 			Config
	shutdownChan	chan struct{}
}

type Config struct {
	URL 			string
	ReconnectDelay 	time.Duration
	MaxReconnect 	int
}

type QueueConfig struct {
	Name 		string
	Durable 	bool
	AutoDelete 	bool
	Exclusive 	bool
	NoWait 		bool
	Args 		amqp.Table
}

type PublishConfig struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Message    amqp.Publishing
}

type ConsumeConfig struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}


func DefaultQueueConfig(name string) QueueConfig {
	return QueueConfig{
		Name: name,
		Durable: true,
		AutoDelete: false,
		Exclusive: false,
		NoWait: false,
		Args: nil,
	}
}

func DefaultPublishConfig(exchange, key string) PublishConfig {
	return PublishConfig{
		Exchange: exchange,
		RoutingKey: key,
		Mandatory: false,
		Immediate: false,
		Message: amqp.Publishing{
			ContentType: "application/json",
		},
	}
}

func DefaultConfig() Config {
	return Config{
		URL:           "amqp://guest:guest@localhost:5672/",
		ReconnectDelay: 5 * time.Second,
		MaxReconnect:   5,
	}
}

func NewRabbitMQ(config Config) (*RabbitMQ, error) {

	if config.URL == "" {
		config = DefaultConfig()
	}

	r := &RabbitMQ {
		config: config,
		shutdownChan: make(chan struct{}),
	}

	if err := r.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Start connection monitor in background
	go r.MonitorConnection()

	return r, nil
}

func (r *RabbitMQ) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error

	for i := 0; i <= r.config.MaxReconnect; i++ {
		if i > 0 {
			log.Printf("Reconnection attempt %d/%d", i, r.config.MaxReconnect)
			time.Sleep(r.config.ReconnectDelay)
		}

		conn, err := amqp.Dial(r.config.URL)

		if err != nil {
			lastErr = err
			// fmt.Errorf("Failed to connect: %w", err)
			continue
		}
		
		channel, err := conn.Channel()
		
		if err != nil {
			conn.Close()
			lastErr = err
			// fmt.Errorf("Failed to connect: %w", err)
			continue
		}

		r.conn = conn
		r.channel = channel
		r.isConnected = true

		log.Println("Successfully connected to RabbitMQ")
		return nil
		
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", r.config.MaxReconnect+1, lastErr)
}

func (r *RabbitMQ) DeclareQueue(config QueueConfig)  (amqp.Queue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isConnected || r.channel == nil {
		return amqp.Queue{}, errors.New("Not connected to RabbitMQ")
	}
	
	queue, err := r.channel.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)

	if err != nil {
		return amqp.Queue{}, fmt.Errorf("Unable to declare queue %w", err)
	}

	log.Printf("Queue declared: %s (messages: %d, consumers: %d)", 
		queue.Name, queue.Messages, queue.Consumers)

	return queue, nil
}

func (r *RabbitMQ) MonitorConnection() {
	notifyClose := r.conn.NotifyClose(make(chan *amqp.Error))

		for {
			select {
				case err := <- notifyClose:
					if err != nil {
						r.mu.Lock()
						r.isConnected = false
						r.mu.Unlock()
						
						if err := r.Connect(); err != nil {
							log.Printf("Failed to reconnect: %s", err.Error())
						}
					}
			case <-r.shutdownChan:
				return
			}
		}
}

func (r *RabbitMQ) Publish(ctx context.Context, config PublishConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isConnected || r.channel == nil {
		return errors.New("Not connected to RabbitMQ")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()
	}

	err := r.channel.PublishWithContext(
		ctx,
		config.Exchange,
		config.RoutingKey,
		config.Mandatory,
		config.Immediate,
		config.Message,
	)

	if err != nil {
		return fmt.Errorf("Failed to publish message: %w", err)
	}

	log.Printf("Message published to %s/%s", config.Exchange, config.RoutingKey)

	return nil
}

func (r *RabbitMQ) Consume (config ConsumeConfig) (<- chan amqp.Delivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isConnected || r.channel == nil {
		return nil, errors.New("Not connected to RabbitMQ")
	}

	msgs, err := r.channel.Consume(
		config.Queue,
		config.Consumer,
		config.AutoAck,
		config.Exclusive,
		config.NoLocal,
		config.NoWait,
		config.Args,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to start consuming: %w", err)
	}

	log.Printf("Consumer started on queue: %s", config.Queue)

	return msgs, nil
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	close(r.shutdownChan)

	var errs []error

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	r.isConnected = false

	if len(errs) > 0 {
		return fmt.Errorf("failed to close RabbitMQ: %v", errs)
	}

	log.Println("RabbitMQ connection closed gracefully")

	return nil
}