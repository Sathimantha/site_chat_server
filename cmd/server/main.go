package main

import (
	"log"
	"os"

	"github.com/Sathimantha/site-chat-server/internal/chat"
	"github.com/Sathimantha/site-chat-server/internal/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize SQLite database
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Get Google API key
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		apiKey = "YOUR_GOOGLE_API_KEY_HERE"
		log.Println("Warning: Using fallback API key (not secure for production)")
	}

	// Set up Gin router
	r := gin.Default()
	r.Use(cors.Default())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Rio Chat Server is Online (Go version)")
	})

	// Register chat API routes
	chat.RegisterRoutes(r, apiKey)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5004"
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
