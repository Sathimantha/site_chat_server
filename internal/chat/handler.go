package chat

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Sathimantha/site-chat-server/internal/db"
	"github.com/gin-gonic/gin"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var systemPrompt string

func init() {
	systemPromptPath := os.Getenv("SYSTEM_PROMPT_PATH")
	if systemPromptPath == "" {
		systemPromptPath = "system_prompt.md"
		log.Println("Warning: Using default system prompt path")
	}
	data, err := os.ReadFile(systemPromptPath)
	if err != nil {
		systemPrompt = "[System prompt could not be loaded]"
		log.Printf("Failed to load system prompt from %s: %v", systemPromptPath, err)
	} else {
		systemPrompt = string(data)
	}
}

func RegisterRoutes(r *gin.Engine, apiKey string) {
	r.POST("/api/chat", handleChat(apiKey))
	r.GET("/api/history", handleHistory)
	r.GET("/api/stats", handleStats)
	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})
}

func handleChat(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Message   string `json:"message" binding:"required"`
			SessionID string `json:"sessionId,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request format: " + err.Error(),
			})
			return
		}

		userMsg := strings.TrimSpace(req.Message)
		if userMsg == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty message"})
			return
		}

		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = "anon_" + c.ClientIP()
		}

		// ────────────────────────────────────────────────
		//          Google Gemini API call
		// ────────────────────────────────────────────────

		ctx := context.Background()

		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Printf("Failed to create Gemini client: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to initialize AI service",
			})
			return
		}
		defer client.Close()

		model := client.GenerativeModel("gemini-flash-latest")

		// Generation settings
		model.SetTemperature(0.9)
		model.SetTopP(0.95)
		model.SetTopK(64)
		model.SetMaxOutputTokens(2048)

		// Safety settings
		model.SafetySettings = []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockMediumAndAbove,
			},
		}

		// Prompt construction
		fullPrompt := systemPrompt + "\n\nUser: " + userMsg + "\n\nAssistant:"

		resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
		if err != nil {
			log.Printf("Gemini GenerateContent failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "AI service error: " + err.Error(),
			})
			return
		}

		// Extract reply text
		var replyBuilder strings.Builder
		for _, cand := range resp.Candidates {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					replyBuilder.WriteString(string(txt))
				}
			}
		}

		reply := strings.TrimSpace(replyBuilder.String())
		if reply == "" {
			reply = "I don't have a response right now. Please try to contact our office: https://wa.me/94771414496"
		}

		// Save messages (no token counts)
		_ = db.SaveMessage(sessionID, "user", userMsg, 0, 0)
		_ = db.SaveMessage(sessionID, "model", reply, 0, 0)

		log.Printf("[reply] session=%s  (first 70 chars: %q)",
			sessionID, reply[:min(70, len(reply))])

		c.JSON(http.StatusOK, gin.H{
			"reply": reply,
		})
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
