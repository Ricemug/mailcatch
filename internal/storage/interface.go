package storage

import "fakesmtp/internal/models"

// Storage defines the interface for email storage implementations
type Storage interface {
	SaveEmail(*models.Email) error
	GetEmails(int) ([]*models.EmailSummary, error)
	GetEmail(int) (*models.Email, error)
	DeleteEmail(int) error
	ClearEmails() error
	GetEmailCount() (int, error)
	Close() error
}