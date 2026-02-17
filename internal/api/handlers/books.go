package handlers

import (
	"encoding/json"
	"errors"
	"libraryapi/internal/api/dto"
	"libraryapi/internal/api/responses"
	"libraryapi/internal/domain/models"
	"libraryapi/internal/domain/repositories"
	"libraryapi/internal/pkg/cache"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type BookHandler struct {
	repo  repositories.BookRepository
	cache cache.Cache
}

func NewBookHandler(repo repositories.BookRepository, cache cache.Cache) *BookHandler {
	return &BookHandler{
		repo:  repo,
		cache: cache,
	}
}

func (h *BookHandler) BooksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetBooks(w, r)
	case http.MethodPost:
		h.AddBook(w, r)
	default:
		responses.MethodNotAllowed(w)
	}
}

func (h *BookHandler) BookByIDHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		responses.BadRequest(w, errors.New("invalid URL: book ID required"))
		return
	}

	id := parts[len(parts)-1]
	if id == "" {
		responses.BadRequest(w, errors.New("book ID cannot be empty"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetBookByID(w, r, id)
	case http.MethodPut, http.MethodPatch:
		h.UpdateBook(w, r, id)
	case http.MethodDelete:
		h.DeleteBook(w, r, id)
	default:
		responses.MethodNotAllowed(w)
	}
}

// extra functions for Getbooks w pagination

type paginatedresponse struct {
	Data []models.Book      `json:"data"`
	Meta dto.PaginationInfo `json:"meta"`
}

func hasfilters(queryparams map[string]string) bool {
	params := []string{"author", "year", "title"}
	for _, param := range params {
		if _, ok := queryparams[param]; ok {
			return true
		}
	}
	return false
}
func generatePaginationCacheKey(pagination dto.Pagination) string {
	k := strconv.Itoa(pagination.Page)
	j := strconv.Itoa(pagination.Limit)
	return "books:page:" + k + ":limit:" + j
}
func calculateTotalPages(totalItems, perPage int) int {
	if perPage == 0 {
		return 0
	}
	if totalItems == 0 {
		return 1
	}

	pages := totalItems / perPage
	if totalItems%perPage > 0 {
		pages++
	}
	return pages
}
func (h *BookHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	queryparams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if k != "" && v[0] != "" && len(v) > 0 {
			queryparams[k] = v[0]
		}
	}
	pagination := dto.Newpaginationfromrequest(queryparams)
	if err := pagination.Validate(); err != nil {
		responses.BadRequest(w, err)
		return
	}

	cacheKey := ""
	if !hasfilters(queryparams) {
		cacheKey = generatePaginationCacheKey(pagination)
		var cachedresponse paginatedresponse
		if err := h.cache.Get(cacheKey, &cachedresponse); err == nil {
			if err := responses.Success(w, cachedresponse, ""); err != nil {
				log.Error().Err(err).Msg("Failed to send cached response")
				if err := h.cache.Delete(cacheKey); err != nil {
					log.Error().Err(err).Msg("Failed to delete cache")
				}
			} else {
				log.Debug().Str("cache_key", cacheKey).Msg("Cache hit for paginated books")
				return
			}
		}
	}
	books, totalItems, err := h.repo.Getall(pagination)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get books from repository")
		responses.InternalError(w, errors.New("failed to get books"))
		return
	}
	if len(books) == 0 {
		responses.NotFound(w, errors.New("no books found"))
		log.Warn().Msg("No books found")
		return
	}
	totalpages := calculateTotalPages(totalItems, pagination.Limit)
	response := paginatedresponse{
		Data: books,
		Meta: dto.PaginationInfo{
			CurrentPage: pagination.Page,
			PerPage:     pagination.Limit,
			TotalPages:  totalpages,
			TotalItems:  totalItems,
			HasNext:     pagination.Page < totalpages,
			HasPrev:     pagination.Page > 1,
		},
	}
	if cacheKey != "" {
		if err := h.cache.Set(cacheKey, books, 5*time.Minute); err != nil {
			log.Warn().Err(err).Msg("Failed to cache books")
		}
	}
	if err := responses.Success(w, response, ""); err != nil {
		log.Error().Err(err).Msg("Failed to send response")
		return
	}
}

func (h *BookHandler) AddBook(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to decode request body")
		responses.BadRequest(w, errors.New("invalid JSON format"))
		return
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Msg("Validation failed for create book request")
		responses.BadRequest(w, err)
		return
	}

	book := h.repo.Create(req.Title, req.Author, req.Year)
	if err := h.cache.Delete("books:all"); err != nil {
		log.Warn().Err(err).Msg("Failed to invalidate cache")
	}

	log.Info().
		Str("book_id", book.ID).
		Str("title", book.Title).
		Msg("Book created")

	if err := responses.Success(w, book, "Book created successfully"); err != nil {
		log.Error().Err(err).Msg("Failed to send create book response")
	}
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request, id string) {
	cacheKey := "book:" + id
	var cachedBook models.Book

	if err := h.cache.Get(cacheKey, &cachedBook); err == nil {
		log.Debug().Str("cache_key", cacheKey).Str("book_id", id).Msg("Cache hit for book")

		if err := responses.Success(w, cachedBook, ""); err != nil {
			log.Error().Err(err).Msg("Failed to send cached book response")
		}
		return
	}

	book, err := h.repo.Getbyid(id)
	if err != nil {
		log.Warn().Str("book_id", id).Err(err).Msg("Book not found")
		responses.NotFound(w, errors.New("book not found"))
		return
	}

	if err := h.cache.Set(cacheKey, book, 10*time.Minute); err != nil {
		log.Warn().Err(err).Str("cache_key", cacheKey).Msg("Failed to cache book")
	}

	if err := responses.Success(w, book, ""); err != nil {
		log.Error().Err(err).Msg("Failed to send book response")
	}
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request, id string) {
	var req dto.UpdateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to decode request body")
		responses.BadRequest(w, errors.New("invalid JSON format"))
		return
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Msg("Validation failed for update book request")
		responses.BadRequest(w, err)
		return
	}

	existingBook, err := h.repo.Getbyid(id)
	if err != nil {
		log.Warn().Str("book_id", id).Err(err).Msg("Book not found for update")
		responses.NotFound(w, errors.New("book not found"))
		return
	}

	updated := false
	if req.Title != nil && *req.Title != existingBook.Title {
		existingBook.Title = *req.Title
		updated = true
	}
	if req.Author != nil && *req.Author != existingBook.Author {
		existingBook.Author = *req.Author
		updated = true
	}
	if req.Year != nil && *req.Year != existingBook.Year {
		existingBook.Year = *req.Year
		updated = true
	}

	if !updated {
		responses.BadRequest(w, errors.New("no changes provided"))
		return
	}

	updatedBook, err := h.repo.Update(id, existingBook)
	if err != nil {
		log.Error().Err(err).Str("book_id", id).Msg("Failed to update book")
		responses.InternalError(w, errors.New("failed to update book"))
		return
	}

	// Invalidate caches
	cacheKeys := []string{"books:all", "book:" + id}
	for _, key := range cacheKeys {
		if err := h.cache.Delete(key); err != nil {
			log.Warn().Err(err).Str("cache_key", key).Msg("Failed to invalidate cache")
		}
	}

	log.Info().Str("book_id", id).Msg("Book updated")

	if err := responses.Success(w, updatedBook, "Book updated successfully"); err != nil {
		log.Error().Err(err).Msg("Failed to send update book response")
	}
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.repo.Delete(id); err != nil {
		log.Warn().Str("book_id", id).Err(err).Msg("Book not found for deletion")
		responses.NotFound(w, errors.New("book not found"))
		return
	}

	// Invalidate caches
	cacheKeys := []string{"books:all", "book:" + id}
	for _, key := range cacheKeys {
		if err := h.cache.Delete(key); err != nil {
			log.Warn().Err(err).Str("cache_key", key).Msg("Failed to invalidate cache")
		}
	}

	log.Info().Str("book_id", id).Msg("Book deleted")

	if err := responses.Success(w, nil, "Book deleted successfully"); err != nil {
		log.Error().Err(err).Msg("Failed to send delete book response")
	}
}

func (h *BookHandler) AddTestBooks() {
	books := []struct {
		title  string
		author string
		year   int
	}{
		{"1984", "George Orwell", 1949},
		{"Animal Farm", "George Orwell", 1945},
		{"Brave New World", "Aldous Huxley", 1932},
		{"To Kill a Mockingbird", "Harper Lee", 1960},
		{"The Great Gatsby", "F. Scott Fitzgerald", 1925},
	}

	for _, b := range books {
		h.repo.Create(b.title, b.author, b.year)
	}

	log.Info().Int("count", len(books)).Msg("Added test books")
}
