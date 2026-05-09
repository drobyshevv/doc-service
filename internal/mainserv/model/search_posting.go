package model

import "github.com/google/uuid"

type SearchPosting struct {
	ID                int64     `db:"id"`
	TermID            int64     `db:"term_id"`
	DocumentID        uuid.UUID `db:"document_id"`
	TermFrequency     int       `db:"term_frequency"`
	DocumentFrequency int       `db:"document_frequency"`
}
