package router

import (
	"pivote/internal/domains/auth"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/program"
	"pivote/internal/domains/user"
	"pivote/internal/domains/vote"
	"pivote/internal/infra/rabbitmq"

	"github.com/gin-gonic/gin"
)

func SetupRouter(mq *rabbitmq.RabbitMQ) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		auth.RegisterRoutes(authRoutes, mq)
	}
	
	{
		userRoutes := v1.Group("/users")
		user.RegisterRoutes(userRoutes)
	}

	{	
		programRoutes := v1.Group("/programs")
		program.RegisterRoutes(programRoutes)
	}

	{
		candidateRoutes := v1.Group("/candidates")
		candidate.RegisterRoutes(candidateRoutes)
	}

	{
		voteRoutes := v1.Group("/votes")
		vote.RegisterRoutes(voteRoutes)
	}

	router.GET("/health", healthCheck)

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hale and healthy",
		"status":  "success",
		"data": gin.H{
			"version": "1.0.0",
		},
	})
}
