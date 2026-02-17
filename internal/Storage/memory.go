/*package Storage

import (
	"errors"
	"libraryapi/internal/api/dto"
	"libraryapi/internal/domain/models"
	"libraryapi/internal/domain/repositories"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Memorystorage struct {
	mu    sync.Mutex
	books map[string]models.Book
}

func NewMemory() repositories.BookRepository {
	return &Memorystorage{
		books: make(map[string]models.Book),
	}
}
func (m *Memorystorage) Getall(pagi dto.Pagination) ([]models.Book, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	totalitems := len(m.books)
	if totalitems == 0 {
		return []models.Book{}, 0, nil
	}
	allbooks := make([]models.Book, 0, totalitems)
	for _, book := range m.books {
		allbooks = append(allbooks, book)
	}
	pagination := pagi.Offset()
	if pagination > totalitems {
		return []models.Book{}, 0, errors.New("pagination out of range")
	} else if pagination+pagi.Limit > totalitems {
		return allbooks[pagination:], totalitems, nil
	} else {
		return allbooks[pagination : pagination+pagi.Limit], totalitems, nil
	}
}
func (m *Memorystorage) Create(title string, author string, year int) models.Book {
	m.mu.Lock()
	defer m.mu.Unlock()
	book := models.Book{
		ID:         uuid.New().String(),
		Title:      title,
		Author:     author,
		Year:       year,
		Created_at: time.Now(),
	}
	m.books[book.ID] = book
	return book
}
func (m *Memorystorage) Getbyid(id string) (models.Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	book, exists := m.books[id]
	if !exists {
		return models.Book{}, errors.New("book not found")
	}
	return book, nil

}
func (m *Memorystorage) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.books[id]; !exists {
		return errors.New("book not found")
	}

	delete(m.books, id)
	return nil
}
func (m *Memorystorage) Update(id string, updated models.Book) (models.Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.books[id]; !exists {
		return models.Book{}, errors.New("book not found")
	}
	updated.ID = id
	updated.Created_at = m.books[id].Created_at
	updated.UpdatedAt = time.Now()
	m.books[id] = updated
	return updated, nil
}

func (m *Memorystorage) Search(title, author string, year int) ([]models.Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []models.Book
	for _, book := range m.books {
		if (title == "" || containsIgnoreCase(book.Title, title)) &&
			(author == "" || containsIgnoreCase(book.Author, author)) &&
			(year == 0 || book.Year == year) {
			result = append(result, book)
		}
	}
	return result, nil
}

func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}*/
