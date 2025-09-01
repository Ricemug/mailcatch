package models

import "time"

type Email struct {
	ID        int       `json:"id" db:"id"`
	From      string    `json:"from" db:"from_addr"`
	To        string    `json:"to" db:"to_addr"`
	Subject   string    `json:"subject" db:"subject"`
	Body      string    `json:"body" db:"body"`
	HTML      string    `json:"html" db:"html"`
	Raw       string    `json:"raw" db:"raw"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type EmailSummary struct {
	ID        int       `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
}