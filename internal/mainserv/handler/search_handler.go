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

// SearchPhrase обрабатывает HTTP-запрос
// на поиск точной фразы.
//
// Используется для поиска документов,
// содержащих последовательность слов в заданном порядке.
func (h *SearchHandler) SearchPhrase(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query().Get("query")

	results, err := h.searchService.SearchPhrase(
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

// SearchPhraseByOwner обрабатывает HTTP-запрос на поиск точной фразы
// среди документов конкретного пользователя.
func (h *SearchHandler) SearchPhraseByOwner(
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

	results, err := h.searchService.SearchPhraseByOwner(
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
