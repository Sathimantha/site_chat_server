package chat

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Sathimantha/site-chat-server/internal/db"
	"github.com/gin-gonic/gin"
)

var systemPrompt string

func init() {
	data, err := os.ReadFile("system_prompt.txt")
	if err != nil {
		systemPrompt = "[System prompt could not be loaded]"
	} else {
		systemPrompt = string(data)
	}
}

func RegisterRoutes(r *gin.Engine, apiKey string) {
	r.POST("/api/chat", handleChat(apiKey))
	r.GET("/api/history", handleHistory)
	r.GET("/api/stats", handleStats)
}

func handleChat(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Message   string `json:"message" binding:"required"`
			SessionID string `json:"sessionId,omitempty"`
		}

		userMsg := strings.TrimSpace(req.Message)
		if userMsg == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty message"})
			return
		}

		reply := "[Gemini SDK call missing: update this block with correct API usage]"

		// Save assistant reply (tokens=0 for now)
		if err := db.SaveMessage(req.SessionID, "model", reply, 0, 0); err != nil {
			log.Printf("save model msg failed: %v", err)
		}

		log.Printf("[reply] session=%s (first 70 chars: %q)", req.SessionID, reply[:min(70, len(reply))])

		c.JSON(http.StatusOK, gin.H{
			"reply": reply,
		})
		// ...existing code...
	}
}

func handleHistory(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusOK, []map[string]string{})
		return
	}

	history, err := db.GetHistory(sessionID)
	if err != nil {
		log.Printf("history fetch failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load history"})
		return
	}

	c.JSON(http.StatusOK, history)
}

func handleStats(c *gin.Context) {
	var totalMessages, totalInput, totalOutput int

	err := db.DB.QueryRow(`
		SELECT 
			COUNT(*)               AS total_messages,
			COALESCE(SUM(input_tokens),  0) AS total_input,
			COALESCE(SUM(output_tokens), 0) AS total_output
		FROM messages
	`).Scan(&totalMessages, &totalInput, &totalOutput)

	if err != nil {
		log.Printf("stats query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_requests": totalMessages,
		"input_tokens":   totalInput,
		"output_tokens":  totalOutput,
		"estimated_cost": "Check Google AI Studio dashboard",
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
