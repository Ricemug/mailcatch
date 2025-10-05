package web

import (
	"net/http"
	"strconv"

	"mailcatch/internal/models"
	"mailcatch/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	storage storage.Storage
	hub     *WebSocketHub
}

func NewHandler(storage storage.Storage, hub *WebSocketHub) *Handler {
	return &Handler{
		storage: storage,
		hub:     hub,
	}
}

func (h *Handler) GetEmails(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	
	emails, err := h.storage.GetEmails(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, emails)
}

func (h *Handler) GetEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}
	
	email, err := h.storage.GetEmail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}
	
	c.JSON(http.StatusOK, email)
}

func (h *Handler) DeleteEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}
	
	err = h.storage.DeleteEmail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Email deleted successfully"})
}

func (h *Handler) ClearEmails(c *gin.Context) {
	err := h.storage.ClearEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "All emails cleared"})
}

func (h *Handler) GetStats(c *gin.Context) {
	count, err := h.storage.GetEmailCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	stats := gin.H{
		"total_emails": count,
		"connected_clients": h.hub.GetClientCount(),
	}
	
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}

func (h *Handler) OnNewEmail(email *models.Email) {
	// Store email in database
	err := h.storage.SaveEmail(email)
	if err != nil {
		return
	}
	
	// Broadcast to all connected WebSocket clients
	summary := &models.EmailSummary{
		ID:        email.ID,
		From:      email.From,
		To:        email.To,
		Subject:   email.Subject,
		CreatedAt: email.CreatedAt,
	}
	
	h.hub.Broadcast("new_email", summary)
}