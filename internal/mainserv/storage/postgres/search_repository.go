package postgres

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drobyshevv/doc-service/internal/mainserv/model"
)

// SearchRepository предоставляет методы
// для работы с поисковым индексом документов.
type SearchRepository struct {
	db    *pgxpool.Pool
	cache *termCache
}

// NewSearchRepository создаёт новый поисковый репозиторий.
func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{
		db:    db,
		cache: newTermCache(),
	}
}

// termCache хранит соответствие term -> termID
// в памяти приложения для уменьшения количества запросов к БД.
type termCache struct {
	mu    sync.RWMutex
	store map[string]int64
}

func newTermCache() *termCache {
	return &termCache{
		store: make(map[string]int64),
	}
}

func (c *termCache) Get(term string) (int64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	id, ok := c.store[term]
	return id, ok
}

func (c *termCache) Set(term string, id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[term] = id
}

// CreateTerm создаёт новый термин в таблице terms.
//
// Если термин уже существует,
// увеличивает document_frequency на 1.
//
// Возвращает ID термина.
func (r *SearchRepository) CreateTerm(
	ctx context.Context,
	term string,
) (int64, error) {

	if id, ok := r.cache.Get(term); ok {
		return id, nil
	}

	query := `
		INSERT INTO terms (term, document_frequency)
		VALUES ($1, 1)
		ON CONFLICT (term)
		DO UPDATE SET document_frequency = terms.document_frequency + 1
		RETURNING id
	`

	var termID int64

	err := r.db.QueryRow(ctx, query, term).Scan(&termID)
	if err != nil {
		return 0, err
	}

	r.cache.Set(term, termID)

	return termID, nil
}

// CreatePosting создаёт связь между термином и документом,
// сохраняя term frequency и TF-IDF score.
func (r *SearchRepository) CreatePosting(
	ctx context.Context,
	posting *model.Posting,
) (int64, error) {

	query := `
		INSERT INTO postings (
			term_id,
			document_id,
			term_frequency
		)
		VALUES ($1,$2,$3)
		RETURNING id
	`

	var postingID int64

	err := r.db.QueryRow(
		ctx,
		query,
		posting.TermID,
		posting.DocumentID,
		posting.TermFrequency,
	).Scan(&postingID)

	if err != nil {
		return 0, err
	}

	return postingID, nil
}

// CreateTermPositions сохраняет позиции термина внутри документа.
//
// Использует batch insert для уменьшения количества запросов к БД.
func (r *SearchRepository) CreateTermPositions(
	ctx context.Context,
	postingID int64,
	positions []int,
) error {

	if len(positions) == 0 {
		return nil
	}

	query := `
		INSERT INTO term_positions (
			posting_id,
			position
		)
		VALUES ($1,$2)
	`

	batch := &pgx.Batch{}

	for _, pos := range positions {
		batch.Queue(query, postingID, pos)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range positions {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// SearchByTerm ищет документы по термину.
//
// Результаты сортируются по TF-IDF score в порядке убывания.
// Возвращается не более 100 записей.
func (r *SearchRepository) SearchByTerm(
	ctx context.Context,
	term string,
) ([]model.SearchPosting, error) {

	query := `
		SELECT
			p.id,
			p.term_id,
			p.document_id,
			p.term_frequency,
			t.document_frequency
		FROM postings p
		JOIN terms t ON p.term_id = t.id
		WHERE t.term = $1
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.SearchPosting

	for rows.Next() {
		var sp model.SearchPosting

		err := rows.Scan(
			&sp.ID,
			&sp.TermID,
			&sp.DocumentID,
			&sp.TermFrequency,
			&sp.DocumentFrequency,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, sp)
	}

	return res, rows.Err()
}

// SuggestTerms возвращает список терминов,
// начинающихся с указанного префикса.
//
// Результаты сортируются по document_frequency,
// что позволяет показывать наиболее популярные подсказки.
func (r *SearchRepository) SuggestTerms(
	ctx context.Context,
	prefix string,
	limit int,
) ([]string, error) {

	query := `
		SELECT term
		FROM terms
		WHERE term ILIKE $1
		ORDER BY document_frequency DESC
		LIMIT $2
	`

	rows, err := r.db.Query(
		ctx,
		query,
		prefix+"%",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []string

	for rows.Next() {
		var term string

		if err := rows.Scan(&term); err != nil {
			return nil, err
		}

		terms = append(terms, term)
	}

	return terms, rows.Err()
}

func (r *SearchRepository) CountDocuments(ctx context.Context) (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM documents`

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SearchRepository) GetPositionsByPosting(
	ctx context.Context,
	postingID int64,
) ([]int, error) {

	query := `
		SELECT position
		FROM term_positions
		WHERE posting_id = $1
		ORDER BY position
	`

	rows, err := r.db.Query(ctx, query, postingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []int

	for rows.Next() {
		var pos int

		if err := rows.Scan(&pos); err != nil {
			return nil, err
		}

		positions = append(positions, pos)
	}

	return positions, rows.Err()
}

// GetPositionsByPostings возвращает позиции для нескольких posting_ids одним запросом.
// Возвращает map[postingID][]position.
func (r *SearchRepository) GetPositionsByPostings(
	ctx context.Context,
	postingIDs []int64,
) (map[int64][]int, error) {
	if len(postingIDs) == 0 {
		return make(map[int64][]int), nil
	}

	query := `
		SELECT posting_id, position
		FROM term_positions
		WHERE posting_id = ANY($1)
		ORDER BY posting_id, position
	`

	rows, err := r.db.Query(ctx, query, postingIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]int)
	for rows.Next() {
		var postingID int64
		var pos int
		if err := rows.Scan(&postingID, &pos); err != nil {
			return nil, err
		}
		result[postingID] = append(result[postingID], pos)
	}

	return result, rows.Err()
}

// SearchByTitle ищет документы по названию или имени файла.
// Использует pg_trgm индекс для быстрого ILIKE.
func (r *SearchRepository) SearchByTitle(
	ctx context.Context,
	query string,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	pattern := "%" + query + "%"

	var sql string
	var args []interface{}

	if userID == uuid.Nil {
		// Аноним: только публичные
		sql = `
			SELECT id FROM documents
			WHERE (title ILIKE $1 OR original_filename ILIKE $1)
			  AND is_public = true
			ORDER BY 
				CASE WHEN title ILIKE $2 THEN 1 ELSE 2 END,
				created_at DESC
			LIMIT $3
		`
		args = []interface{}{pattern, query, 100}
	} else {
		// Авторизованный: свои + публичные
		sql = `
			SELECT id FROM documents
			WHERE (title ILIKE $1 OR original_filename ILIKE $1)
			  AND (owner_id = $2 OR is_public = true)
			ORDER BY 
				CASE WHEN title ILIKE $3 THEN 1 ELSE 2 END,
				created_at DESC
			LIMIT $4
		`
		args = []interface{}{pattern, userID, query, 100}
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
