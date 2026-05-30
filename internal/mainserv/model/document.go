package model

import (
	"time"

	uuid "github.com/google/uuid"
)

type Document struct {
	ID               uuid.UUID `db:"id"`
	OwnerID          uuid.UUID `db:"owner_id"`
	Title            string    `db:"title"`
	OriginalFilename string    `db:"original_filename"`
	S3Key            string    `db:"s3_key"`
	IsPublic         bool      `db:"is_public"`
	FileSize         int64     `db:"file_size"`
	MimeType         string    `db:"mime_type"`
	TokenCount       int       `db:"token_count"`
	CurrentVersion   int       `db:"current_version"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type DocumentVersion struct {
	ID         uuid.UUID `db:"id" json:"id"`
	DocumentID uuid.UUID `db:"document_id" json:"document_id"`
	Version    int       `db:"version" json:"version"`
	S3Key      string    `db:"s3_key" json:"s3_key"`
	FileSize   int64     `db:"file_size" json:"file_size"`
	MimeType   string    `db:"mime_type" json:"mime_type"`
	UploadedBy uuid.UUID `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	Note       string    `db:"note" json:"note"`
}

// UpdateMetadataInput содержит поля для обновления метаданных документа.
// Все поля опциональны (указываются только те, которые нужно изменить).
type UpdateMetadataInput struct {
	Title    *string `json:"title,omitempty"`
	IsPublic *bool   `json:"is_public,omitempty"`
}

// ListDocumentsQuery содержит параметры фильтрации и пагинации.
type ListDocumentsQuery struct {
	Page      int
	Limit     int
	SortBy    string // "created_at", "title", "file_size"
	SortOrder string // "asc", "desc"
	MimeType  *string
	IsPublic  *bool
	From      *time.Time
	To        *time.Time
	Search    *string // поиск по title или original_filename
}

// DocumentMeta содержит только метаданные документа (без содержимого файла).
type DocumentMeta struct {
	ID               uuid.UUID `json:"id" db:"id"`
	OwnerID          uuid.UUID `json:"owner_id" db:"owner_id"`
	Title            string    `json:"title" db:"title"`
	OriginalFilename string    `json:"original_filename" db:"original_filename"`
	IsPublic         bool      `json:"is_public" db:"is_public"`
	FileSize         int64     `json:"file_size" db:"file_size"`
	MimeType         string    `json:"mime_type" db:"mime_type"`
	TokenCount       int       `json:"token_count" db:"token_count"`
	CurrentVersion   int       `json:"current_version" db:"current_version"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
