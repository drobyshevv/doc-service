package model

type Term struct {
	ID                int64  `db:"id"`
	Term              string `db:"term"`
	DocumentFrequency int    `db:"document_frequency"`
}
