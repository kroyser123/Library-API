package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"libraryapi/internal/api/dto"
	"libraryapi/internal/domain/models"
	"libraryapi/internal/domain/repositories"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgres(connectionString string) (repositories.BookRepository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Устанавливаем настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStorage{db: db}, nil
}

// GetAll получает книги с пагинацией
func (p *PostgresStorage) Getall(pagination dto.Pagination) ([]models.Book, int, error) {
	// 1. Получаем общее количество книг
	var totalItems int
	err := p.db.QueryRow("SELECT COUNT(*) FROM books").Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count books: %w", err)
	}

	// 2. Если нет книг - возвращаем пустой список
	if totalItems == 0 {
		return []models.Book{}, 0, nil
	}

	// 3. Получаем книги с пагинацией
	offset := pagination.Offset()
	query := `
		SELECT id, title, author, year, created_at, updated_at 
		FROM books 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := p.db.Query(query, pagination.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query books: %w", err)
	}
	defer rows.Close()

	// 4. Сканируем результаты
	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.Year,
			&book.Created_at,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan book: %w", err)
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return books, totalItems, nil
}

// Getbyid получает книгу по ID
func (p *PostgresStorage) Getbyid(id string) (models.Book, error) {
	query := `
		SELECT id, title, author, year, created_at, updated_at 
		FROM books 
		WHERE id = $1
	`

	var book models.Book
	err := p.db.QueryRow(query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
		&book.Created_at,
		&book.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Book{}, errors.New("book not found")
		}
		return models.Book{}, fmt.Errorf("failed to get book: %w", err)
	}

	return book, nil
}

// Create создает новую книгу
func (p *PostgresStorage) Create(title string, author string, year int) models.Book {
	query := `
		INSERT INTO books (id, title, author, year, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, author, year, created_at, updated_at
	`

	var book models.Book
	id := uuid.New().String()

	err := p.db.QueryRow(query, id, title, author, year, time.Now()).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
		&book.Created_at,
		&book.UpdatedAt,
	)

	if err != nil {
		// В случае ошибки, возвращаем книгу с сгенерированным ID
		// В реальном проекте нужно обрабатывать ошибку
		return models.Book{
			ID:         id,
			Title:      title,
			Author:     author,
			Year:       year,
			Created_at: time.Now(),
		}
	}

	return book
}

// Update обновляет книгу
func (p *PostgresStorage) Update(id string, updated models.Book) (models.Book, error) {
	query := `
		UPDATE books 
		SET title = $1, author = $2, year = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, title, author, year, created_at, updated_at
	`

	var book models.Book
	err := p.db.QueryRow(
		query,
		updated.Title,
		updated.Author,
		updated.Year,
		time.Now(),
		id,
	).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
		&book.Created_at,
		&book.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Book{}, errors.New("book not found")
		}
		return models.Book{}, fmt.Errorf("failed to update book: %w", err)
	}

	return book, nil
}

// Delete удаляет книгу
func (p *PostgresStorage) Delete(id string) error {
	query := "DELETE FROM books WHERE id = $1"
	result, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("book not found")
	}

	return nil
}

// Search ищет книги по фильтрам
func (p *PostgresStorage) Search(title, author string, year int) ([]models.Book, error) {
	query := `
		SELECT id, title, author, year, created_at, updated_at 
		FROM books 
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
			AND ($2 = '' OR author ILIKE '%' || $2 || '%')
			AND ($3 = 0 OR year = $3)
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(query, title, author, year)
	if err != nil {
		return nil, fmt.Errorf("failed to search books: %w", err)
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.Year,
			&book.Created_at,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return books, nil
}
