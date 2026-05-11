package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(conn *sql.DB) *DB {
	return &DB{Conn: conn}
}
