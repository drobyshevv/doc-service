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
