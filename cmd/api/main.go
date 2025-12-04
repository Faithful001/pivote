package main

import (
	"pivote/internal/db"
	"pivote/internal/domains/user"
	"pivote/internal/router"
)

func main() {
	// Initialize database connection
	db.InitDB()

	// Run migrations
	if err := db.AutoMigrate(&user.User{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Setup router
	r := router.SetupRouter()

	// Start server
	r.Run(":8000")
}