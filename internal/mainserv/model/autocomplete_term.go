package model

type AutocompleteTerm struct {
	ID        int64  `db:"id"`
	Term      string `db:"term"`
	Frequency int    `db:"frequency"`
}
