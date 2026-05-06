package model

import uuid "github.com/google/uuid"

type Posting struct {
	ID            int64     `db:"id"`
	TermID        int64     `db:"term_id"`
	DocumentID    uuid.UUID `db:"document_id"`
	TermFrequency int       `db:"term_frequency"`
	TFIDFScore    float64   `db:"tfidf_score"`
}
