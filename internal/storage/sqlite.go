package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"fakesmtp/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

func (s *SQLiteStorage) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS emails (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_addr TEXT NOT NULL,
		to_addr TEXT NOT NULL,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		html TEXT,
		raw TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_emails_created_at ON emails(created_at DESC);
	`
	
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStorage) SaveEmail(email *models.Email) error {
	query := `
		INSERT INTO emails (from_addr, to_addr, subject, body, html, raw, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := s.db.Exec(query, email.From, email.To, email.Subject, 
		email.Body, email.HTML, email.Raw, email.CreatedAt)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	email.ID = int(id)
	return nil
}

func (s *SQLiteStorage) GetEmails(limit int) ([]*models.EmailSummary, error) {
	query := `
		SELECT id, from_addr, to_addr, subject, created_at 
		FROM emails 
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var emails []*models.EmailSummary
	for rows.Next() {
		email := &models.EmailSummary{}
		err := rows.Scan(&email.ID, &email.From, &email.To, 
			&email.Subject, &email.CreatedAt)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	
	return emails, nil
}

func (s *SQLiteStorage) GetEmail(id int) (*models.Email, error) {
	query := `
		SELECT id, from_addr, to_addr, subject, body, html, raw, created_at
		FROM emails 
		WHERE id = ?
	`
	
	email := &models.Email{}
	err := s.db.QueryRow(query, id).Scan(
		&email.ID, &email.From, &email.To, &email.Subject,
		&email.Body, &email.HTML, &email.Raw, &email.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return email, nil
}

func (s *SQLiteStorage) DeleteEmail(id int) error {
	query := `DELETE FROM emails WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *SQLiteStorage) ClearEmails() error {
	query := `DELETE FROM emails`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStorage) GetEmailCount() (int, error) {
	query := `SELECT COUNT(*) FROM emails`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	return count, err
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}