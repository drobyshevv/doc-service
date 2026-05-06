package storage

import "errors"

var (
	// ErrDocumentNotFound возвращается,
	// когда документ не найден в хранилище.
	ErrDocumentNotFound = errors.New("document not found")
)
