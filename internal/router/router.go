package router

import (
	"pivote/internal/domains/auth"
	"pivote/internal/domains/candidate"
	"pivote/internal/domains/otp"
	"pivote/internal/domains/program"
	"pivote/internal/domains/user"
	"pivote/internal/domains/vote"
	"pivote/internal/domains/workspace"
	"pivote/internal/infra/rabbitmq"
	"pivote/internal/infra/sse"
	"pivote/internal/infra/websocket"
	"pivote/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter(mq *rabbitmq.RabbitMQ, socketio *websocket.SocketIOServer, sseBroadcaster *sse.BroadcasterManager) *gin.Engine {
	router := gin.Default()

	router.Use(middlewares.CORS())

	v1 := router.Group("/api/v1")

	{
		authRoutes := v1.Group("/auth")
		auth.RegisterRoutes(authRoutes, mq)
	}

	{
		otpRoutes := v1.Group("/otps")
		otp.RegisterRoutes(otpRoutes, mq)
	}

	{
		userAuth := v1.Group("/users")
		userAuth.Use(middlewares.Standard(user.RoleAdmin, user.RoleUser))

		userAdmin := v1.Group("/users")
		userAdmin.Use(middlewares.Standard(user.RoleAdmin))

		user.RegisterRoutes(userAuth, userAdmin)
	}

	{
		programRoutes := v1.Group("/programs")
		program.RegisterRoutes(programRoutes, mq, sseBroadcaster)
	}

	{
		candidateRoutes := v1.Group("/candidates")
		candidate.RegisterRoutes(candidateRoutes)
	}

	{
		voteRoutes := v1.Group("/votes")
		vote.RegisterRoutes(voteRoutes, socketio)
	}

	{
		workspace.RegisterRoutes(v1)
	}

	router.GET("/socket.io/*any", gin.WrapH(socketio.Server))
	router.POST("/socket.io/*any", gin.WrapH(socketio.Server))

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
