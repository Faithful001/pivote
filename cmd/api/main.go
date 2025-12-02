package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/health", func (c *gin.Context){
		c.JSON(http.StatusOK, gin.H{
			"message": "Hale and healthy",
			"status": "success",
			"data": gin.H{
				"version": "1.0.0",
			},
		})
	})

	router.Run(":8000")
}