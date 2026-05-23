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

// GetPostingByDocumentAndTerm возвращает posting
// для конкретного документа и термина.
func (r *SearchRepository) GetPostingByDocumentAndTerm(
	ctx context.Context,
	documentID uuid.UUID,
	termID int64,
) (*model.Posting, error) {

	query := `
		SELECT
			id,
			term_id,
			document_id,
			term_frequency,
		FROM postings
		WHERE document_id = $1 AND term_id = $2
	`

	var posting model.Posting

	err := r.db.QueryRow(ctx, query, documentID, termID).Scan(
		&posting.ID,
		&posting.TermID,
		&posting.DocumentID,
		&posting.TermFrequency,
	)
	if err != nil {
		return nil, err
	}

	return &posting, nil
}

// SearchPhrase выполняет поиск документов,
// содержащих набор терминов поисковой фразы.
//
// Метод используется как основа
// для реализации phrase search.
func (r *SearchRepository) SearchPhrase(
	ctx context.Context,
	terms []string,
) ([]model.SearchPosting, error) {

	query := `
		SELECT DISTINCT
			p.id,
			p.term_id,
			p.document_id,
			p.term_frequency,
			t.document_frequency
		FROM postings p
		JOIN terms t ON p.term_id = t.id
		WHERE t.term = ANY($1)
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, terms)
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

func (r *SearchRepository) GetDocumentFrequency(
	ctx context.Context,
	term string,
) (int, error) {

	query := `
		SELECT document_frequency
		FROM terms
		WHERE term = $1
	`

	var df int

	err := r.db.QueryRow(ctx, query, term).Scan(&df)
	if err != nil {
		return 0, err
	}

	return df, nil
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
