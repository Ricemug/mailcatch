package config

import (
	"flag"
	"os"
	"path/filepath"
)

type Config struct {
	SMTPPort        string
	HTTPPort        string
	DBPath          string
	LogPath         string
	ClearOnShutdown bool
	Daemon          bool
}

func Load() *Config {
	cfg := &Config{}
	
	// Default log path to temp directory
	defaultLogPath := filepath.Join(os.TempDir(), "mailcatch.log")
	
	flag.StringVar(&cfg.SMTPPort, "smtp-port", "2525", "SMTP server port")
	flag.StringVar(&cfg.HTTPPort, "http-port", "8080", "HTTP server port")
	flag.StringVar(&cfg.DBPath, "db-path", "./data/emails.db", "Database file path")
	flag.StringVar(&cfg.LogPath, "log-path", defaultLogPath, "Log file path (default: temp directory)")
	flag.BoolVar(&cfg.ClearOnShutdown, "clear-on-shutdown", true, "Clear all emails when shutting down")
	flag.BoolVar(&cfg.Daemon, "daemon", false, "Run in background as daemon")
	flag.Parse()

	// Environment variables override flags
	if port := os.Getenv("SMTP_PORT"); port != "" {
		cfg.SMTPPort = port
	}
	if port := os.Getenv("HTTP_PORT"); port != "" {
		cfg.HTTPPort = port
	}
	if path := os.Getenv("DB_PATH"); path != "" {
		cfg.DBPath = path
	}
	if path := os.Getenv("LOG_PATH"); path != "" {
		cfg.LogPath = path
	}
	if clear := os.Getenv("CLEAR_ON_SHUTDOWN"); clear == "false" {
		cfg.ClearOnShutdown = false
	}
	if daemon := os.Getenv("DAEMON"); daemon == "true" {
		cfg.Daemon = true
	}

	return cfg
}