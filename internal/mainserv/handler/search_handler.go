package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/drobyshevv/doc-service/internal/mainserv/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(
	searchService *service.SearchService,
) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search обрабатывает HTTP-запрос
// на выполнение полнотекстового поиска.
//
// Выполняет поиск документов по query parameter q
// и возвращает отсортированный список результатов.
func (h *SearchHandler) Search(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query().Get("query")

	results, err := h.searchService.Search(
		r.Context(),
		query,
		20,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(results)
}

// SearchByOwner обрабатывает HTTP-запрос
// на поиск документов конкретного пользователя.
//
// Выполняет полнотекстовый поиск
// только среди документов указанного owner_id.
func (h *SearchHandler) SearchByOwner(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query().Get("query")
	owner := r.URL.Query().Get("owner_id")

	ownerID, err := uuid.Parse(owner)
	if err != nil {
		http.Error(w, "invalid owner_id", http.StatusBadRequest)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user header", http.StatusInternalServerError)
		return
	}

	if ownerID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	results, err := h.searchService.SearchByOwner(
		r.Context(),
		ownerID,
		query,
		20,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(results)
}

// Suggest обрабатывает HTTP-запрос
// на получение поисковых подсказок.
//
// Возвращает список терминов,
// начинающихся с указанного префикса.
func (h *SearchHandler) Suggest(
	w http.ResponseWriter,
	r *http.Request,
) {
	prefix := r.URL.Query().Get("query")

	results, err := h.searchService.Suggest(
		r.Context(),
		prefix,
		10,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(results)
}

// SearchByTitle обрабатывает поиск по названию или имени файла.
// Возвращает те же поля, что и обычный Search.
func (h *SearchHandler) SearchByTitle(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query().Get("query")
	if query == "" {
		json.NewEncoder(w).Encode([]service.SearchResult{})
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	var userID uuid.UUID
	if userIDStr != "" {
		var err error
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "invalid user header", http.StatusInternalServerError)
			return
		}
	}

	results, err := h.searchService.SearchByTitle(
		r.Context(),
		query,
		20,
		userID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(results)
}
