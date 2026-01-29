package router

import (
	"libraryapi/internal/api/handlers"
	"libraryapi/internal/api/middleware"
	"net/http"
)

func SetupRouter(bookHandler *handlers.BookHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Library API v1.0"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/api/books", bookHandler.BooksHandler)
	mux.HandleFunc("/api/books/", bookHandler.BookByIDHandler)

	// Apply middleware chain: Recovery -> Logger
	return middleware.Chain(
		middleware.Recovery,
		middleware.Logger,
	)(mux)
}
