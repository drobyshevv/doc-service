package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/drobyshevv/doc-service/internal/mainserv/model"
	"github.com/drobyshevv/doc-service/internal/mainserv/search/tokenizer"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage/postgres"
	redisstorage "github.com/drobyshevv/doc-service/internal/mainserv/storage/redis"
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
	redis      *redisstorage.Client
}

type SearchResult struct {
	Document *model.Document `json:"document"`
	Score    float64         `json:"score"`
}

func NewSearchService(
	searchRepo *postgres.SearchRepository,
	docRepo *postgres.DocumentRepository,
	redisClient *redisstorage.Client,
) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		docRepo:    docRepo,
		redis:      redisClient,
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

/*
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
*/

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

	cacheKey := fmt.Sprintf("search:%s:%d", query, limit)

	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil && len(cached) > 0 {
		var results []SearchResult
		if json.Unmarshal([]byte(cached), &results) == nil {
			return results, nil
		}
	}

	terms := tokenizer.Tokenize(query)
	terms = uniqueTerms(terms)

	if len(terms) > maxTerms {
		terms = terms[:maxTerms]
	}

	totalDocs, err := s.searchRepo.CountDocuments(ctx)
	if err != nil {
		return nil, err
	}

	documentScores := make(map[uuid.UUID]float64)
	docSet := make(map[uuid.UUID]struct{})

	for _, term := range terms {
		postings, err := s.searchRepo.SearchByTerm(ctx, term)
		if err != nil {
			return nil, err
		}

		for _, p := range postings {
			docSet[p.DocumentID] = struct{}{}

			score := calcScore(
				p.TermFrequency,
				p.DocumentFrequency,
				totalDocs,
				0,
			)

			documentScores[p.DocumentID] += score
		}
	}

	docIDs := make([]uuid.UUID, 0, len(docSet))
	for id := range docSet {
		docIDs = append(docIDs, id)
	}

	docs, err := s.docRepo.GetByIDs(ctx, docIDs)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(documentScores))

	for id, score := range documentScores {
		doc := docs[id]
		if doc == nil {
			continue
		}

		if doc.TokenCount > 0 {
			score = score / math.Sqrt(float64(doc.TokenCount))
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

	bytes, _ := json.Marshal(results)
	err = s.redis.Set(ctx, cacheKey, bytes, 5*time.Minute)
	if err != nil {
		log.Println("REDIS SET ERROR:", err)
	}

	return results, nil
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

	terms := tokenizer.Tokenize(query)
	if len(terms) < 2 {
		return s.Search(ctx, query, limit)
	}

	limit = normalizeLimit(limit)

	totalDocs, err := s.searchRepo.CountDocuments(ctx)
	if err != nil {
		return nil, err
	}

	postingsByTerm := make([][]model.SearchPosting, len(terms))

	for i, term := range terms {
		postings, err := s.searchRepo.SearchByTerm(ctx, term)
		if err != nil {
			return nil, err
		}
		postingsByTerm[i] = postings
	}

	docMap := make(map[uuid.UUID][]model.SearchPosting)

	for _, postings := range postingsByTerm {
		for _, p := range postings {
			docMap[p.DocumentID] = append(
				docMap[p.DocumentID],
				p,
			)
		}
	}

	documentScores := make(map[uuid.UUID]float64)

	for docID, postings := range docMap {

		if len(postings) != len(terms) {
			continue
		}

		var allPositions [][]int

		for _, p := range postings {

			positions, err := s.searchRepo.GetPositionsByPosting(
				ctx,
				p.ID,
			)
			if err != nil {
				return nil, err
			}

			allPositions = append(
				allPositions,
				positions,
			)
		}

		if !isPhraseMatch(allPositions) {
			continue
		}

		score := 0.0

		for _, p := range postings {
			score += calcScore(
				p.TermFrequency,
				p.DocumentFrequency,
				totalDocs,
				0,
			)
		}

		documentScores[docID] = score
	}

	docIDs := make([]uuid.UUID, 0, len(documentScores))
	for id := range documentScores {
		docIDs = append(docIDs, id)
	}

	docs, err := s.docRepo.GetByIDs(ctx, docIDs)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0)

	for id, score := range documentScores {

		doc := docs[id]
		if doc == nil {
			continue
		}

		if doc.TokenCount > 0 {
			score /= math.Sqrt(
				float64(doc.TokenCount),
			)
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

func calcScore(
	tf int,
	docsWithTerm int,
	totalDocs int,
	docLength int,
) float64 {

	tfNorm := 1 + math.Log(float64(tf))

	idf := math.Log(
		(float64(totalDocs) + 1) /
			(float64(docsWithTerm) + 1),
	)

	score := tfNorm * idf

	if docLength > 0 {
		score = score / math.Sqrt(float64(docLength))
	}

	return score
}

func uniqueTerms(terms []string) []string {
	seen := make(map[string]struct{})
	res := make([]string, 0, len(terms))

	for _, t := range terms {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		res = append(res, t)
	}

	if len(res) > maxTerms {
		return res[:maxTerms]
	}

	return res
}

func isPhraseMatch(
	allPositions [][]int,
) bool {

	if len(allPositions) == 0 {
		return false
	}

	first := allPositions[0]

	for _, start := range first {

		match := true
		current := start

		for i := 1; i < len(allPositions); i++ {

			found := false

			for _, pos := range allPositions[i] {
				if pos == current+1 {
					found = true
					current = pos
					break
				}
			}

			if !found {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}
