package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"mailcatch/internal/models"
	"go.etcd.io/bbolt"
)

type BoltStorage struct {
	db *bbolt.DB
}

var (
	emailsBucket = []byte("emails")
	metaBucket   = []byte("meta")
	nextIDKey    = []byte("next_id")
)

func NewBoltStorage(dbPath string) (*BoltStorage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Change extension to .bolt
	boltPath := dbPath[:len(dbPath)-len(filepath.Ext(dbPath))] + ".bolt"

	db, err := bbolt.Open(boltPath, 0644, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt database: %w", err)
	}

	storage := &BoltStorage{db: db}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(emailsBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(metaBucket)
		if err != nil {
			return err
		}
		
		// Initialize next ID if it doesn't exist
		metaBucket := tx.Bucket(metaBucket)
		if metaBucket.Get(nextIDKey) == nil {
			return metaBucket.Put(nextIDKey, []byte("1"))
		}
		return nil
	})

	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return storage, nil
}

func (s *BoltStorage) SaveEmail(email *models.Email) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		emails := tx.Bucket(emailsBucket)
		meta := tx.Bucket(metaBucket)

		// Get next ID
		nextIDBytes := meta.Get(nextIDKey)
		nextID, err := strconv.Atoi(string(nextIDBytes))
		if err != nil {
			return err
		}

		email.ID = nextID

		// Serialize email
		data, err := json.Marshal(email)
		if err != nil {
			return err
		}

		// Store email
		key := []byte(strconv.Itoa(nextID))
		if err := emails.Put(key, data); err != nil {
			return err
		}

		// Update next ID
		nextID++
		return meta.Put(nextIDKey, []byte(strconv.Itoa(nextID)))
	})
}

func (s *BoltStorage) GetEmails(limit int) ([]*models.EmailSummary, error) {
	var summaries []*models.EmailSummary

	err := s.db.View(func(tx *bbolt.Tx) error {
		emails := tx.Bucket(emailsBucket)
		
		// Collect all emails first to sort by timestamp
		var allEmails []*models.Email
		
		c := emails.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var email models.Email
			if err := json.Unmarshal(v, &email); err != nil {
				continue // Skip corrupted entries
			}
			allEmails = append(allEmails, &email)
		}

		// Sort by created time descending
		sort.Slice(allEmails, func(i, j int) bool {
			return allEmails[i].CreatedAt.After(allEmails[j].CreatedAt)
		})

		// Apply limit and create summaries
		if limit > len(allEmails) {
			limit = len(allEmails)
		}

		summaries = make([]*models.EmailSummary, limit)
		for i := 0; i < limit; i++ {
			email := allEmails[i]
			summaries[i] = &models.EmailSummary{
				ID:        email.ID,
				From:      email.From,
				To:        email.To,
				Subject:   email.Subject,
				CreatedAt: email.CreatedAt,
			}
		}

		return nil
	})

	return summaries, err
}

func (s *BoltStorage) GetEmail(id int) (*models.Email, error) {
	var email models.Email

	err := s.db.View(func(tx *bbolt.Tx) error {
		emails := tx.Bucket(emailsBucket)
		key := []byte(strconv.Itoa(id))
		
		data := emails.Get(key)
		if data == nil {
			return fmt.Errorf("email not found")
		}

		return json.Unmarshal(data, &email)
	})

	if err != nil {
		return nil, err
	}

	return &email, nil
}

func (s *BoltStorage) DeleteEmail(id int) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		emails := tx.Bucket(emailsBucket)
		key := []byte(strconv.Itoa(id))
		
		if emails.Get(key) == nil {
			return fmt.Errorf("email not found")
		}

		return emails.Delete(key)
	})
}

func (s *BoltStorage) ClearEmails() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		// Delete the bucket and recreate it
		if err := tx.DeleteBucket(emailsBucket); err != nil {
			return err
		}
		
		_, err := tx.CreateBucket(emailsBucket)
		return err
	})
}

func (s *BoltStorage) GetEmailCount() (int, error) {
	var count int

	err := s.db.View(func(tx *bbolt.Tx) error {
		emails := tx.Bucket(emailsBucket)
		stats := emails.Stats()
		count = stats.KeyN
		return nil
	})

	return count, err
}

func (s *BoltStorage) Close() error {
	return s.db.Close()
}