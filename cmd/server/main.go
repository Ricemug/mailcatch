package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"fakesmtp/internal/config"
	"fakesmtp/internal/models"
	"fakesmtp/internal/smtp"
	"fakesmtp/internal/storage"
	"fakesmtp/internal/web"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Setup logging
	if err := setupLogging(cfg.LogPath, cfg.Daemon); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}
	
	// Handle daemon mode
	if cfg.Daemon {
		log.Printf("Starting FakeSMTP in daemon mode...")
		log.Printf("SMTP: %s, HTTP: %s, DB: %s", cfg.SMTPPort, cfg.HTTPPort, cfg.DBPath)
		log.Printf("Log file: %s", cfg.LogPath)
	}
	
	// Initialize storage (try SQLite first, fallback to BoltDB)
	var storageInstance storage.Storage
	
	sqliteStorage, err := storage.NewSQLiteStorage(cfg.DBPath)
	if err != nil {
		log.Printf("SQLite not available (CGO disabled), trying BoltDB: %v", err)
		boltStorage, err := storage.NewBoltStorage(cfg.DBPath)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
		storageInstance = boltStorage
		log.Println("Using BoltDB storage")
	} else {
		storageInstance = sqliteStorage
		log.Println("Using SQLite storage")
	}
	defer storageInstance.Close()
	
	// Initialize web server
	webServer := web.NewServer(storageInstance)
	
	// Initialize SMTP server with email handler
	smtpServer := smtp.NewServer(cfg.SMTPPort, func(email *models.Email) {
		webServer.GetEmailHandler()(email)
	})
	
	// Start servers
	go func() {
		log.Printf("Starting web server on port %s", cfg.HTTPPort)
		if err := webServer.Start(cfg.HTTPPort); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()
	
	go func() {
		log.Printf("Starting SMTP server on port %s", cfg.SMTPPort)
		if err := smtpServer.Start(); err != nil {
			log.Fatalf("Failed to start SMTP server: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down servers...")
	smtpServer.Stop()
	
	// Clear all test emails on shutdown (if enabled)
	if cfg.ClearOnShutdown {
		log.Println("Clearing all test emails...")
		if err := storageInstance.ClearEmails(); err != nil {
			log.Printf("Error clearing emails: %v", err)
		} else {
			log.Println("All test emails cleared")
		}
	}
	
	log.Println("Servers stopped")
}

func setupLogging(logPath string, daemon bool) error {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	if daemon {
		// In daemon mode, only log to file
		log.SetOutput(logFile)
	} else {
		// In interactive mode, log to both stdout and file
		multiWriter := &MultiWriter{
			writers: []interface{}{os.Stdout, logFile},
		}
		log.SetOutput(multiWriter)
	}
	
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

// MultiWriter implements io.Writer to write to multiple destinations
type MultiWriter struct {
	writers []interface{}
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		if writer, ok := w.(interface{ Write([]byte) (int, error) }); ok {
			n, err = writer.Write(p)
			if err != nil {
				return
			}
		}
	}
	return len(p), nil
}