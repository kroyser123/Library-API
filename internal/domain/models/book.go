package models

import "time"

type Book struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Year       int       `json:"year"`
	Created_at time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}