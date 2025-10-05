package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"mailcatch/internal/models"
)

type MemoryStorage struct {
	emails   []*models.Email
	nextID   int
	mutex    sync.RWMutex
	filePath string
}

func NewMemoryStorage(dbPath string) (*MemoryStorage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Change extension to .json
	jsonPath := dbPath[:len(dbPath)-len(filepath.Ext(dbPath))] + ".json"
	
	storage := &MemoryStorage{
		emails:   make([]*models.Email, 0),
		nextID:   1,
		filePath: jsonPath,
	}

	// Load existing data if file exists
	if err := storage.loadFromFile(); err != nil {
		// If file doesn't exist, it's not an error
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load data: %w", err)
		}
	}

	return storage, nil
}

func (s *MemoryStorage) loadFromFile() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var fileData struct {
		NextID int              `json:"next_id"`
		Emails []*models.Email  `json:"emails"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return err
	}

	s.emails = fileData.Emails
	s.nextID = fileData.NextID

	return nil
}

func (s *MemoryStorage) saveToFile() error {
	fileData := struct {
		NextID int              `json:"next_id"`
		Emails []*models.Email  `json:"emails"`
	}{
		NextID: s.nextID,
		Emails: s.emails,
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *MemoryStorage) SaveEmail(email *models.Email) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	email.ID = s.nextID
	s.nextID++
	s.emails = append(s.emails, email)

	// Keep only last 1000 emails to prevent unlimited growth
	if len(s.emails) > 1000 {
		s.emails = s.emails[len(s.emails)-1000:]
	}

	return s.saveToFile()
}

func (s *MemoryStorage) GetEmails(limit int) ([]*models.EmailSummary, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Sort by created time descending
	emailsCopy := make([]*models.Email, len(s.emails))
	copy(emailsCopy, s.emails)
	
	sort.Slice(emailsCopy, func(i, j int) bool {
		return emailsCopy[i].CreatedAt.After(emailsCopy[j].CreatedAt)
	})

	// Apply limit
	if limit > len(emailsCopy) {
		limit = len(emailsCopy)
	}

	summaries := make([]*models.EmailSummary, limit)
	for i := 0; i < limit; i++ {
		email := emailsCopy[i]
		summaries[i] = &models.EmailSummary{
			ID:        email.ID,
			From:      email.From,
			To:        email.To,
			Subject:   email.Subject,
			CreatedAt: email.CreatedAt,
		}
	}

	return summaries, nil
}

func (s *MemoryStorage) GetEmail(id int) (*models.Email, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, email := range s.emails {
		if email.ID == id {
			return email, nil
		}
	}

	return nil, fmt.Errorf("email not found")
}

func (s *MemoryStorage) DeleteEmail(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, email := range s.emails {
		if email.ID == id {
			s.emails = append(s.emails[:i], s.emails[i+1:]...)
			return s.saveToFile()
		}
	}

	return fmt.Errorf("email not found")
}

func (s *MemoryStorage) ClearEmails() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.emails = make([]*models.Email, 0)
	return s.saveToFile()
}

func (s *MemoryStorage) GetEmailCount() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.emails), nil
}

func (s *MemoryStorage) Close() error {
	return s.saveToFile()
}