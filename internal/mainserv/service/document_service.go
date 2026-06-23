package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime"
	"path/filepath"

	"github.com/drobyshevv/doc-service/internal/mainserv/extractor"
	"github.com/drobyshevv/doc-service/internal/mainserv/model"
	"github.com/drobyshevv/doc-service/internal/mainserv/search/indexer"
	"github.com/drobyshevv/doc-service/internal/mainserv/search/tokenizer"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage/postgres"
	redisstorage "github.com/drobyshevv/doc-service/internal/mainserv/storage/redis"
	s3storage "github.com/drobyshevv/doc-service/internal/mainserv/storage/s3"
	"github.com/google/uuid"
)

type DocumentService struct {
	docRepo    *postgres.DocumentRepository
	searchRepo *postgres.SearchRepository
	s3Storage  *s3storage.Storage
	redis      *redisstorage.Client
}

func NewDocumentService(
	docRepo *postgres.DocumentRepository,
	searchRepo *postgres.SearchRepository,
	s3Storage *s3storage.Storage,
	redisClient *redisstorage.Client,
) *DocumentService {
	return &DocumentService{
		docRepo:    docRepo,
		searchRepo: searchRepo,
		s3Storage:  s3Storage,
		redis:      redisClient,
	}
}

type UploadDocumentInput struct {
	OwnerID  uuid.UUID
	Title    string
	Filename string
	Data     []byte
	IsPublic bool
}

func (s *DocumentService) UploadDocument(
	ctx context.Context,
	input UploadDocumentInput,
) (*model.Document, error) {

	docID := uuid.New()
	s3Key := s3storage.GenerateKey(input.Filename)

	contentType := mime.TypeByExtension(filepath.Ext(input.Filename))
	if contentType == "" {
		contentType = "text/plain"
	}

	err := s.s3Storage.UploadFile(ctx, s3Key, input.Data, contentType)
	if err != nil {
		return nil, err
	}

	title := input.Title
	if title == "" {
		title = input.Filename
	}

	ext := extractor.NewExtractor(contentType, input.Filename)
	fileReader := bytes.NewReader(input.Data)
	extractedText, err := ext.Extract(fileReader, contentType)
	log.Printf("[DEBUG] Filename: %s, MIME: %s", input.Filename, contentType)
	log.Printf("[DEBUG] Extracted text (first 200 chars): %q",
		func() string {
			if len(extractedText) > 200 {
				return extractedText[:200] + "..."
			}
			return extractedText
		}())
	if err != nil {
		log.Printf("[WARN] Failed to extract text from %s: %v. Indexing metadata only.", input.Filename, err)
		extractedText = ""
	}

	tokens := tokenizer.Tokenize(extractedText)

	log.Printf("[DEBUG] Tokens count: %d, first 10: %v", len(tokens),
		func() []string {
			if len(tokens) > 10 {
				return tokens[:10]
			}
			return tokens
		}())

	doc := &model.Document{
		ID:               docID,
		OwnerID:          input.OwnerID,
		Title:            title,
		OriginalFilename: input.Filename,
		S3Key:            s3Key,
		IsPublic:         input.IsPublic,
		FileSize:         int64(len(input.Data)),
		MimeType:         contentType,
		TokenCount:       len(tokens),
		CurrentVersion:   1,
	}

	err = s.docRepo.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	version := &model.DocumentVersion{
		DocumentID: docID,
		Version:    1,
		S3Key:      s3Key,
		FileSize:   doc.FileSize,
		MimeType:   doc.MimeType,
		UploadedBy: input.OwnerID,
		Note:       "Initial version",
	}
	if err := s.docRepo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	err = s.indexDocument(ctx, docID, extractedText)
	if err != nil {
		return nil, err
	}

	s.invalidateSearchCache(ctx, extractedText)

	return doc, nil
}

func (s *DocumentService) indexDocument(
	ctx context.Context,
	docID uuid.UUID,
	text string,
) error {

	indexedTerms := indexer.BuildIndex(docID, text)

	for _, indexedTerm := range indexedTerms {
		termID, err := s.searchRepo.CreateTerm(
			ctx,
			indexedTerm.Term,
		)
		if err != nil {
			return err
		}

		postingID, err := s.searchRepo.CreatePosting(
			ctx,
			&model.Posting{
				TermID:        termID,
				DocumentID:    docID,
				TermFrequency: indexedTerm.Frequency,
			},
		)
		if err != nil {
			return err
		}

		err = s.searchRepo.CreateTermPositions(ctx, postingID, indexedTerm.Positions)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DocumentService) GetDocument(
	ctx context.Context,
	docID uuid.UUID,
) (*model.Document, []byte, error) {

	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, nil, err
	}

	data, err := s.s3Storage.DownloadFile(ctx, doc.S3Key)
	if err != nil {
		return nil, nil, err
	}

	return doc, data, nil
}

func (s *DocumentService) DeleteDocument(
	ctx context.Context,
	docID uuid.UUID,
) error {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}

	err = s.s3Storage.DeleteFile(ctx, doc.S3Key)
	if err != nil {
		return err
	}

	err = s.docRepo.Delete(ctx, docID)
	if err != nil {
		return err
	}

	return nil
}

// UploadNewVersion загружает новую версию существующего документа.
func (s *DocumentService) UploadNewVersion(
	ctx context.Context,
	docID uuid.UUID,
	uploaderID uuid.UUID,
	data []byte,
	filename string,
	note string,
) (*model.DocumentVersion, error) {

	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.OwnerID != uploaderID {
		return nil, fmt.Errorf("forbidden: not document owner")
	}

	newVersion, err := s.docRepo.GetNextVersionNumber(ctx, docID)
	if err != nil {
		return nil, err
	}

	s3Key := fmt.Sprintf("doc/%s/v%d/%s", docID.String(), newVersion, uuid.New().String()+"-"+filename)

	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "text/plain"
	}

	if err := s.s3Storage.UploadFile(ctx, s3Key, data, contentType); err != nil {
		return nil, err
	}

	version := &model.DocumentVersion{
		DocumentID: docID,
		Version:    newVersion,
		S3Key:      s3Key,
		FileSize:   int64(len(data)),
		MimeType:   contentType,
		UploadedBy: uploaderID,
		Note:       note,
	}
	if err := s.docRepo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	if err := s.docRepo.UpdateCurrentVersion(
		ctx, docID, newVersion, s3Key, filename, contentType, int64(len(data)),
	); err != nil {
		return nil, err
	}

	versionExt := extractor.NewExtractor(contentType, filename)
	versionReader := bytes.NewReader(data)
	versionText, err := versionExt.Extract(versionReader, contentType)
	if err != nil {
		log.Printf("[WARN] Failed to extract text from version of %s: %v", filename, err)
		versionText = ""
	}

	// Индексируем извлечённый текст
	_ = s.indexDocument(ctx, docID, versionText)
	s.invalidateSearchCache(ctx, versionText)

	return version, nil
}

// GetDocumentVersion получает конкретную версию документа.
func (s *DocumentService) GetDocumentVersion(
	ctx context.Context,
	docID uuid.UUID,
	version int,
) (*model.DocumentVersion, []byte, error) {

	ver, err := s.docRepo.GetVersion(ctx, docID, version)
	if err != nil {
		return nil, nil, err
	}

	data, err := s.s3Storage.DownloadFile(ctx, ver.S3Key)
	if err != nil {
		return nil, nil, err
	}

	return ver, data, nil
}

// RollbackToVersion делает указанную версию текущей (rollback).
func (s *DocumentService) RollbackToVersion(
	ctx context.Context,
	docID uuid.UUID,
	targetVersion int,
	uploaderID uuid.UUID,
) error {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.OwnerID != uploaderID {
		return fmt.Errorf("forbidden")
	}

	targetVer, err := s.docRepo.GetVersion(ctx, docID, targetVersion)
	if err != nil {
		return err
	}

	if err := s.docRepo.UpdateCurrentVersion(
		ctx, docID, targetVersion, targetVer.S3Key,
		doc.OriginalFilename, targetVer.MimeType, targetVer.FileSize,
	); err != nil {
		return err
	}

	return nil
}

// ListDocumentVersions получает список версий документа.
func (s *DocumentService) ListDocumentVersions(
	ctx context.Context,
	docID uuid.UUID,
) ([]model.DocumentVersion, error) {
	return s.docRepo.ListVersions(ctx, docID)
}

// UpdateDocumentMetadata обновляет метаданные документа.
func (s *DocumentService) UpdateDocumentMetadata(
	ctx context.Context,
	docID uuid.UUID,
	updaterID uuid.UUID,
	input model.UpdateMetadataInput,
) (*model.Document, error) {

	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.OwnerID != updaterID {
		return nil, fmt.Errorf("forbidden: not document owner")
	}

	updatedDoc, err := s.docRepo.UpdateMetadata(ctx, docID, input)
	if err != nil {
		return nil, err
	}

	if input.IsPublic != nil && *input.IsPublic != doc.IsPublic {
		s.invalidateSearchCacheByDocument(ctx, docID)
	}

	return updatedDoc, nil
}

// ListDocuments возвращает список документов с фильтрацией и пагинацией.
func (s *DocumentService) ListDocuments(
	ctx context.Context,
	userID uuid.UUID,
	query model.ListDocumentsQuery,
) ([]model.Document, int, error) {
	return s.docRepo.ListWithFilters(ctx, userID, query)
}

// GetDocumentMetadata возвращает только метаданные документа (без загрузки файла).
func (s *DocumentService) GetDocumentMetadata(
	ctx context.Context,
	docID uuid.UUID,
	requesterID uuid.UUID,
) (*model.DocumentMeta, error) {

	meta, err := s.docRepo.GetMetadataByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	if !meta.IsPublic && meta.OwnerID != requesterID {
		return nil, fmt.Errorf("forbidden")
	}

	return meta, nil
}

// invalidateSearchCacheByDocument удаляет все ключи кеша, связанные с документом.
// Использует тот же экстрактор, что и при загрузке, чтобы токены совпадали.
func (s *DocumentService) invalidateSearchCacheByDocument(ctx context.Context, docID uuid.UUID) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return
	}

	data, err := s.s3Storage.DownloadFile(ctx, doc.S3Key)
	if err != nil {
		return
	}

	contentType := doc.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(doc.OriginalFilename))
	}
	if contentType == "" {
		contentType = "text/plain"
	}

	ext := extractor.NewExtractor(contentType, doc.OriginalFilename)
	fileReader := bytes.NewReader(data)
	extractedText, err := ext.Extract(fileReader, contentType)
	if err != nil {
		log.Printf("[WARN] Failed to extract text for cache invalidation of %s: %v", doc.OriginalFilename, err)
		extractedText = ""
	}

	s.invalidateSearchCache(ctx, extractedText)
}

// ListByOwner возвращает список документов конкретного пользователя.
func (s *DocumentService) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]model.Document, error) {
	return s.docRepo.ListByOwner(ctx, ownerID)
}

// invalidateSearchCache удаляет ключи кеша поиска по терминам документа.
func (s *DocumentService) invalidateSearchCache(ctx context.Context, text string) {
	terms := tokenizer.Tokenize(text)
	terms = uniqueTermsForCache(terms)

	for _, term := range terms {
		setKey := fmt.Sprintf("cache_keys:term:%s", term)

		// Получаем все ключи кеша, связанные с этим токеном
		keys, err := s.redis.SMembers(ctx, setKey)
		if err != nil {
			log.Printf("cache invalidation error for term %s: %v", term, err)
			continue
		}

		// Удаляем каждый ключ кеша
		for _, key := range keys {
			_ = s.redis.Delete(ctx, key)
		}

		// Удаляем сам SET
		_ = s.redis.Delete(ctx, setKey)
	}
}

func uniqueTermsForCache(terms []string) []string {
	seen := make(map[string]struct{})
	res := make([]string, 0, 5)
	for _, t := range terms {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		res = append(res, t)
		if len(res) >= 5 {
			break
		}
	}
	return res
}
