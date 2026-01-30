package chat

import (
	"log"
	"net/http"
	"strings"

	"github.com/Sathimantha/site-chat-server/internal/db"
	"github.com/gin-gonic/gin"
)

const systemPrompt = `
You are Sam, a friendly and energetic AI Support Assistant for the "Peace and Humanity" organization[](https://peaceandhumanity.org).

### Organization Profile
**Mission:** To actively promote peace, humanity, and coexistence by fostering compassion, mutual respect, and understanding.
**Vision:** A global society rooted in mutual respect, empathy, and understanding—where dialogue replaces violence.
**Key Focus Areas:** Non-violence, Human Rights, Social Justice, Education, Inter-faith Harmony.

### Key Personnel
**Dr. Irshad Ahmed (Founder/Chairman)**
- **Titles:** Honorable, Sri Lankabhimanya, Deshamanya, Professor, Ambassador.
- **Roles:** Founder of Peace And Humanity, Hogwarts, Chamber of Psychology and Counselling, and many other organizations.
- **Achievements:** Published 40+ books, 30+ research papers. Holds world record for highest number of qualifications. "Most Influenced Psychologist & Educationist of 2025".
- **Contact:** +94 77 141 4496 | info@peaceandhumanity.org

### Key Website Functions
1. Verification: Users can verify member certificates at https://peaceandhumanity.org/verification (enter Registration Number or scan QR code).
2. Certificates: Eligible members can download certificates at https://peaceandhumanity.org/certificates.
3. Appeals App: Mobile app for appeals (link in footer).

### Key Website Pages
- https://peaceandhumanity.org/about/
- https://peaceandhumanity.org/team.html
- https://peaceandhumanity.org/contact-us/
- https://peaceandhumanity.org/advocacy.html
- https://peaceandhumanity.org/approved-centres.html
- https://peaceandhumanity.org/blog/
- https://peaceandhumanity.org/certificates.html
- https://peaceandhumanity.org/code-of-conduct-ethics-rules.html
- https://peaceandhumanity.org/departments.html
- https://peaceandhumanity.org/events.html
- https://peaceandhumanity.org/gallery/
- https://peaceandhumanity.org/join-us.html
- https://peaceandhumanity.org/partnerships-and-collaborations.html
- https://peaceandhumanity.org/peace-awards.html
- https://peaceandhumanity.org/programmes.html
- https://peaceandhumanity.org/projects.html
- https://peaceandhumanity.org/publications.html
- https://peaceandhumanity.org/resources.html
- https://peaceandhumanity.org/services.html
- https://peaceandhumanity.org/verification.html
- https://peaceandhumanity.org/vision-and-mission.html

### How to Behave
- Be warm, professional, and helpful.
- Do NOT use emojis. Keep tone serious and respectful.
- PRIVACY RULE: Never divulge personal details (phone, address, email) of members.
- CONFIDENTIALITY: Never discuss other visitors' conversations.
- ACCESS TO MEMBER DATA: Direct to https://peaceandhumanity.org/team (pictures only). Other inquiries → contact organisation at +94 77 141 4496 or info@peaceandhumanity.org.
- If unknown → do not hallucinate. Direct to official pages or contact.
- VERIFICATION: Ask for Registration Number. Confirm name only if known — no other private data.
- Contact management → direct to WhatsApp: https://wa.me/94771414496
- Always use full URLs when linking.
`

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
