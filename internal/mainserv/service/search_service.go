package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/drobyshevv/doc-service/internal/mainserv/model"
	"github.com/drobyshevv/doc-service/internal/mainserv/search/tokenizer"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage/postgres"
)

const (
	defaultLimit   = 20
	maxLimit       = 100
	maxQueryLength = 500
	maxTerms       = 10
)

var (
	ErrQueryTooLong = errors.New("search query is too long")
)

type SearchService struct {
	searchRepo *postgres.SearchRepository
	docRepo    *postgres.DocumentRepository
}

type SearchResult struct {
	Document *model.Document `json:"document"`
	Score    float64         `json:"score"`
}

func NewSearchService(
	searchRepo *postgres.SearchRepository,
	docRepo *postgres.DocumentRepository,
) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		docRepo:    docRepo,
	}
}

// normalizeQuery выполняет предварительную обработку поискового запроса.
//
// Удаляет лишние пробелы и проверяет ограничения
// на максимальную длину запроса.
func normalizeQuery(query string) (string, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return "", nil
	}

	if len(query) > maxQueryLength {
		return "", ErrQueryTooLong
	}

	return query, nil
}

// normalizeLimit валидирует ограничение количества результатов.
//
// Если limit выходит за допустимые пределы,
// возвращается значение по умолчанию.
func normalizeLimit(limit int) int {
	if limit <= 0 || limit > maxLimit {
		return defaultLimit
	}
	return limit
}

// aggregateResults преобразует map documentID -> score
// в отсортированный список SearchResult.
//
// Для каждого документа загружаются метаданные,
// после чего результаты сортируются по релевантности.
func aggregateResults(
	documentScores map[uuid.UUID]float64,
	docRepo *postgres.DocumentRepository,
	ctx context.Context,
	limit int,
) ([]SearchResult, error) {

	results := make([]SearchResult, 0, len(documentScores))

	for docID, score := range documentScores {
		doc, err := docRepo.GetByID(ctx, docID)
		if err != nil {
			continue
		}

		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Search выполняет полнотекстовый поиск документов по поисковому запросу.
//
// Алгоритм работы:
//   - нормализует и валидирует входной запрос;
//   - разбивает запрос на термины;
//   - выполняет поиск postings для каждого термина;
//   - агрегирует TF-IDF score по документам;
//   - сортирует результаты по убыванию релевантности.
//
// Возвращает список документов, наиболее релевантных запросу.
func (s *SearchService) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]SearchResult, error) {

	query, err := normalizeQuery(query)
	if err != nil || query == "" {
		return []SearchResult{}, err
	}

	limit = normalizeLimit(limit)

	terms := tokenizer.Tokenize(query)
	if len(terms) > maxTerms {
		terms = terms[:maxTerms]
	}

	documentScores := make(map[uuid.UUID]float64)

	for _, term := range terms {
		postings, err := s.searchRepo.SearchByTerm(ctx, term)
		if err != nil {
			return nil, err
		}

		for _, posting := range postings {
			documentScores[posting.DocumentID] += posting.TFIDFScore
		}
	}

	return aggregateResults(documentScores, s.docRepo, ctx, limit)
}

// SearchByOwner выполняет полнотекстовый поиск
// только среди документов конкретного пользователя.
//
// Метод использует базовый Search,
// после чего фильтрует результаты по ownerID.
//
// Применяется для поиска по приватным документам пользователя.
func (s *SearchService) SearchByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
	query string,
	limit int,
) ([]SearchResult, error) {

	results, err := s.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	filtered := make([]SearchResult, 0)

	for _, result := range results {
		if result.Document.OwnerID == ownerID {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// SearchPhrase выполняет поиск точной фразы.
//
// В отличие от стандартного поиска,
// учитывает последовательность терминов в запросе
// и возвращает документы, содержащие искомую фразу.
//
// Используется для более точного поиска
// по нескольким связанным словам.
func (s *SearchService) SearchPhrase(
	ctx context.Context,
	query string,
	limit int,
) ([]SearchResult, error) {

	query, err := normalizeQuery(query)
	if err != nil || query == "" {
		return []SearchResult{}, err
	}

	limit = normalizeLimit(limit)

	terms := tokenizer.Tokenize(query)
	if len(terms) < 2 {
		return s.Search(ctx, query, limit)
	}

	documentScores := make(map[uuid.UUID]float64)

	postings, err := s.searchRepo.SearchPhrase(ctx, terms)
	if err != nil {
		return nil, err
	}

	for _, posting := range postings {
		documentScores[posting.DocumentID] += posting.TFIDFScore
	}

	return aggregateResults(documentScores, s.docRepo, ctx, limit)
}

// Suggest возвращает список терминов,
// начинающихся с указанного префикса.
//
// Используется для реализации autocomplete
// и поисковых подсказок при вводе запроса.
func (s *SearchService) Suggest(
	ctx context.Context,
	prefix string,
	limit int,
) ([]string, error) {

	prefix, err := normalizeQuery(prefix)
	if err != nil || prefix == "" {
		return []string{}, err
	}

	limit = normalizeLimit(limit)

	return s.searchRepo.SuggestTerms(ctx, prefix, limit)
}
