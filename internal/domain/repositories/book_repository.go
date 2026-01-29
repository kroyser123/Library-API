package repositories

import (
	"libraryapi/internal/api/dto"
	"libraryapi/internal/domain/models"
)

type BookRepository interface {
	Getall(pagi dto.Pagination) ([]models.Book, int, error)
	Getbyid(id string) (models.Book, error)
	Create(title string, author string, year int) models.Book
	Update(id string, updated models.Book) (models.Book, error)
	Delete(id string) error
	Search(title, author string, year int) ([]models.Book, error)
}
