package web

import (
	"io/fs"
	"net/http"

	"mailcatch/internal/models"
	"mailcatch/internal/storage"
	"github.com/gin-gonic/gin"
)


type Server struct {
	router  *gin.Engine
	handler *Handler
	hub     *WebSocketHub
}

func NewServer(storage storage.Storage) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	
	hub := NewWebSocketHub()
	handler := NewHandler(storage, hub)
	
	server := &Server{
		router:  router,
		handler: handler,
		hub:     hub,
	}
	
	server.setupRoutes()
	go hub.Run()
	
	return server
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/emails", s.handler.GetEmails)
		api.GET("/emails/:id", s.handler.GetEmail)
		api.DELETE("/emails/:id", s.handler.DeleteEmail)
		api.DELETE("/emails", s.handler.ClearEmails)
		api.GET("/stats", s.handler.GetStats)
	}
	
	// WebSocket endpoint
	s.router.GET("/ws", s.handler.HandleWebSocket)
	
	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err == nil {
		s.router.StaticFS("/static", http.FS(staticFS))
	}
	
	// Serve index.html for all other routes (SPA)
	s.router.NoRoute(func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index.html not found")
			return
		}
		c.Data(http.StatusOK, "text/html", data)
	})
}

func (s *Server) Start(port string) error {
	return s.router.Run(":" + port)
}

func (s *Server) GetEmailHandler() func(email interface{}) {
	return func(email interface{}) {
		if e, ok := email.(*models.Email); ok {
			s.handler.OnNewEmail(e)
		}
	}
}

