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
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
