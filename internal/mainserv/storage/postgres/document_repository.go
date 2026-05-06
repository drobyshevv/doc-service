package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drobyshevv/doc-service/internal/mainserv/model"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage"
)

// DocumentRepository предоставляет методы
// для работы с документами в PostgreSQL.
type DocumentRepository struct {
	db *pgxpool.Pool
}

// NewDocumentRepository создаёт новый репозиторий документов.
func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create сохраняет новый документ в базе данных.
//
// После вставки заполняет CreatedAt и UpdatedAt,
// сгенерированные PostgreSQL.
func (r *DocumentRepository) Create(
	ctx context.Context,
	doc *model.Document,
) error {
	query := `
		INSERT INTO documents (
			id,
			owner_id,
			title,
			original_filename,
			s3_key,
			is_public,
			file_size,
			mime_type,
			token_count
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		doc.ID,
		doc.OwnerID,
		doc.Title,
		doc.OriginalFilename,
		doc.S3Key,
		doc.IsPublic,
		doc.FileSize,
		doc.MimeType,
		doc.TokenCount,
	).Scan(
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
}

// GetByID возвращает документ по его ID.
//
// Если документ не найден, возвращает storage.ErrDocumentNotFound.
func (r *DocumentRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Document, error) {
	query := `
		SELECT
			id,
			owner_id,
			title,
			original_filename,
			s3_key,
			is_public,
			file_size,
			mime_type,
			token_count,
			created_at,
			updated_at
		FROM documents
		WHERE id = $1
	`

	var doc model.Document

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.OwnerID,
		&doc.Title,
		&doc.OriginalFilename,
		&doc.S3Key,
		&doc.IsPublic,
		&doc.FileSize,
		&doc.MimeType,
		&doc.TokenCount,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, storage.ErrDocumentNotFound
	}

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

// ListByOwner возвращает список документов пользователя,
// отсортированный по дате создания (новые сверху).
//
// Возвращается не более 100 документов.
func (r *DocumentRepository) ListByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
) ([]model.Document, error) {
	query := `
		SELECT
			id,
			owner_id,
			title,
			original_filename,
			s3_key,
			is_public,
			file_size,
			mime_type,
			created_at,
			updated_at
		FROM documents
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []model.Document

	for rows.Next() {
		var doc model.Document

		err := rows.Scan(
			&doc.ID,
			&doc.OwnerID,
			&doc.Title,
			&doc.OriginalFilename,
			&doc.S3Key,
			&doc.IsPublic,
			&doc.FileSize,
			&doc.MimeType,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		documents = append(documents, doc)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return documents, nil
}

// Delete удаляет документ по ID.
func (r *DocumentRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	query := `
		DELETE FROM documents
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return storage.ErrDocumentNotFound
	}

	return nil
}
