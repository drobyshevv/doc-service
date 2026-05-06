package model

type TermPosition struct {
	ID        int64 `db:"id"`
	PostingID int64 `db:"posting_id"`
	Position  int   `db:"position"`
}
