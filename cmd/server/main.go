package main

import (
	"log"
	"os"
	"strings"

	"github.com/Sathimantha/site-chat-server/internal/chat"
	"github.com/Sathimantha/site-chat-server/internal/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()
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

	// Get allowed origins from env
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsStr != "" {
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}
	if len(allowedOrigins) == 0 {
		log.Println("Warning: No allowed origins specified in ALLOWED_ORIGINS env var. CORS will not allow any origins.")
	}

	// Set up Gin router
	r := gin.Default()
	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))

	// Security best practices
	gin.SetMode(gin.ReleaseMode) // Use release mode in production
	r.SetTrustedProxies(nil)     // Do not trust all proxies

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Rio Chat Server is Online (Go version)")
	})

	// Register chat API routes
	chat.RegisterRoutes(r, apiKey)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}

	certPath := os.Getenv("SSL_CERT_PATH")
	keyPath := os.Getenv("SSL_KEY_PATH")
	if certPath != "" && keyPath != "" {
		log.Printf("Starting HTTPS server on :%s", port)
		if err := r.RunTLS(":"+port, certPath, keyPath); err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		log.Printf("Starting HTTP server on :%s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}
}
