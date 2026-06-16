package redis

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	redisURL := os.Getenv("REDIS_URL")

	var opts *redis.Options

	if redisURL != "" {
		parsed, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("[Redis] Failed to parse REDIS_URL: %v", err)
		}
		opts = parsed
	} else {
		opts = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	Client = redis.NewClient(opts)

	// Verify the connection
	if err := Client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("[Redis] Failed to connect: %v", err)
	}

	fmt.Println("[Redis] Connected successfully")
}