package postgres

import (
	"context"
	"fmt"
	"strings"

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

func (r *DocumentRepository) GetByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) (map[uuid.UUID]*model.Document, error) {

	if len(ids) == 0 {
		return nil, nil
	}

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
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*model.Document)

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
			&doc.TokenCount,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result[doc.ID] = &doc
	}

	return result, rows.Err()
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

// CreateVersion создает новую версию документа.
func (r *DocumentRepository) CreateVersion(
	ctx context.Context,
	version *model.DocumentVersion,
) error {
	query := `
		INSERT INTO document_versions (
			document_id, version, s3_key, file_size,
			mime_type, uploaded_by, note
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		ctx, query,
		version.DocumentID, version.Version, version.S3Key,
		version.FileSize, version.MimeType, version.UploadedBy, version.Note,
	).Scan(&version.ID, &version.CreatedAt)
}

// ListVersions получает список версий документа (от новых к старым).
func (r *DocumentRepository) ListVersions(
	ctx context.Context,
	docID uuid.UUID,
) ([]model.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version, s3_key, file_size,
		       mime_type, uploaded_by, created_at, note
		FROM document_versions
		WHERE document_id = $1
		ORDER BY version DESC
	`
	rows, err := r.db.Query(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []model.DocumentVersion
	for rows.Next() {
		var v model.DocumentVersion
		err := rows.Scan(
			&v.ID, &v.DocumentID, &v.Version, &v.S3Key,
			&v.FileSize, &v.MimeType, &v.UploadedBy,
			&v.CreatedAt, &v.Note,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// UpdateCurrentVersion обновляет текущую версию и метаданные документа.
func (r *DocumentRepository) UpdateCurrentVersion(
	ctx context.Context,
	docID uuid.UUID,
	newVersion int,
	s3Key string,
	filename string,
	mimeType string,
	fileSize int64,
) error {
	query := `
		UPDATE documents 
		SET current_version = $1, 
		    s3_key = $2,
		    original_filename = $3,
		    mime_type = $4,
		    file_size = $5,
		    updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, newVersion, s3Key, filename, mimeType, fileSize, docID)
	return err
}

// GetVersionByS3Key получает версию по S3-ключу (для скачивания).
func (r *DocumentRepository) GetVersionByS3Key(
	ctx context.Context,
	s3Key string,
) (*model.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version, s3_key, file_size,
		       mime_type, uploaded_by, created_at, note
		FROM document_versions WHERE s3_key = $1
	`
	var v model.DocumentVersion
	err := r.db.QueryRow(ctx, query, s3Key).Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.S3Key,
		&v.FileSize, &v.MimeType, &v.UploadedBy,
		&v.CreatedAt, &v.Note,
	)
	if err == pgx.ErrNoRows {
		return nil, storage.ErrDocumentNotFound
	}
	return &v, err
}

// GetVersion получает конкретную версию документа.
func (r *DocumentRepository) GetVersion(
	ctx context.Context,
	docID uuid.UUID,
	version int,
) (*model.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version, s3_key, file_size,
		       mime_type, uploaded_by, created_at, note
		FROM document_versions
		WHERE document_id = $1 AND version = $2
	`
	var v model.DocumentVersion
	err := r.db.QueryRow(ctx, query, docID, version).Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.S3Key,
		&v.FileSize, &v.MimeType, &v.UploadedBy,
		&v.CreatedAt, &v.Note,
	)
	if err == pgx.ErrNoRows {
		return nil, storage.ErrDocumentNotFound
	}
	return &v, err
}

// GetNextVersionNumber возвращает следующий доступный номер версии для документа.
func (r *DocumentRepository) GetNextVersionNumber(ctx context.Context, docID uuid.UUID) (int, error) {
	var maxVer int
	err := r.db.QueryRow(
		ctx,
		"SELECT COALESCE(MAX(version), 0) FROM document_versions WHERE document_id = $1",
		docID,
	).Scan(&maxVer)
	if err != nil {
		return 0, fmt.Errorf("get max version: %w", err)
	}
	return maxVer + 1, nil
}

// UpdateMetadata обновляет метаданные документа (title, is_public).
// Возвращает обновлённый документ с актуальными временными метками.
func (r *DocumentRepository) UpdateMetadata(
	ctx context.Context,
	docID uuid.UUID,
	input model.UpdateMetadataInput,
) (*model.Document, error) {

	// Динамически строим UPDATE-запрос только для указанных полей
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *input.Title)
		argIdx++
	}
	if input.IsPublic != nil {
		updates = append(updates, fmt.Sprintf("is_public = $%d", argIdx))
		args = append(args, *input.IsPublic)
		argIdx++
	}

	if len(updates) == 0 {
		// Ничего не обновляем — возвращаем текущее состояние
		return r.GetByID(ctx, docID)
	}

	// Добавляем updated_at и ID
	updates = append(updates, fmt.Sprintf("updated_at = NOW(), id = $%d", argIdx))
	args = append(args, docID)

	query := fmt.Sprintf(`
		UPDATE documents 
		SET %s
		WHERE id = $%d
		RETURNING 
			id, owner_id, title, original_filename, s3_key, is_public,
			file_size, mime_type, token_count, current_version, created_at, updated_at
	`, strings.Join(updates, ", "), argIdx)

	var doc model.Document
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&doc.ID, &doc.OwnerID, &doc.Title, &doc.OriginalFilename,
		&doc.S3Key, &doc.IsPublic, &doc.FileSize, &doc.MimeType,
		&doc.TokenCount, &doc.CurrentVersion, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, storage.ErrDocumentNotFound
	}
	return &doc, err
}

// ListWithFilters возвращает список документов с поддержкой фильтрации и пагинации.
// Показывает только документы пользователя + публичные документы других пользователей.
func (r *DocumentRepository) ListWithFilters(
	ctx context.Context,
	userID uuid.UUID,
	query model.ListDocumentsQuery,
) ([]model.Document, int, error) {

	// Базовый запрос
	baseQuery := `
		SELECT COUNT(*) OVER() as total,
		       id, owner_id, title, original_filename, s3_key, is_public,
		       file_size, mime_type, token_count, current_version, created_at, updated_at
		FROM documents
		WHERE owner_id = $1 OR is_public = true
	`

	args := []interface{}{userID}
	argIdx := 2

	// Динамические фильтры
	if query.MimeType != nil && *query.MimeType != "" {
		baseQuery += fmt.Sprintf(" AND mime_type = $%d", argIdx)
		args = append(args, *query.MimeType)
		argIdx++
	}
	if query.IsPublic != nil {
		baseQuery += fmt.Sprintf(" AND is_public = $%d", argIdx)
		args = append(args, *query.IsPublic)
		argIdx++
	}
	if query.From != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *query.From)
		argIdx++
	}
	if query.To != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *query.To)
		argIdx++
	}
	if query.Search != nil && *query.Search != "" {
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR original_filename ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+*query.Search+"%")
		argIdx++
	}

	// Сортировка (белый список допустимых полей)
	validSort := map[string]bool{"created_at": true, "title": true, "file_size": true}
	sortBy := query.SortBy
	if sortBy == "" || !validSort[sortBy] {
		sortBy = "created_at"
	}
	sortOrder := "DESC"
	if strings.ToUpper(query.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Пагинация
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Выполняем запрос
	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var documents []model.Document
	var total int

	for rows.Next() {
		var doc model.Document
		err := rows.Scan(
			&total,
			&doc.ID, &doc.OwnerID, &doc.Title, &doc.OriginalFilename,
			&doc.S3Key, &doc.IsPublic, &doc.FileSize, &doc.MimeType,
			&doc.TokenCount, &doc.CurrentVersion, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		documents = append(documents, doc)
	}

	return documents, total, rows.Err()
}

// GetMetadataByID возвращает только метаданные документа (без обращения к S3).
func (r *DocumentRepository) GetMetadataByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.DocumentMeta, error) {
	query := `
		SELECT id, owner_id, title, original_filename, is_public,
		       file_size, mime_type, token_count, current_version, created_at, updated_at
		FROM documents WHERE id = $1
	`

	var meta model.DocumentMeta
	err := r.db.QueryRow(ctx, query, id).Scan(
		&meta.ID, &meta.OwnerID, &meta.Title, &meta.OriginalFilename,
		&meta.IsPublic, &meta.FileSize, &meta.MimeType, &meta.TokenCount,
		&meta.CurrentVersion, &meta.CreatedAt, &meta.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, storage.ErrDocumentNotFound
	}
	return &meta, err
}
