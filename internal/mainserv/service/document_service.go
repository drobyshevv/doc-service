package service

import (
	"context"
	"math"
	"mime"
	"path/filepath"

	"github.com/drobyshevv/doc-service/internal/mainserv/model"
	"github.com/drobyshevv/doc-service/internal/mainserv/search/indexer"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage/postgres"
	s3storage "github.com/drobyshevv/doc-service/internal/mainserv/storage/s3"
	"github.com/google/uuid"
)

type DocumentService struct {
	docRepo    *postgres.DocumentRepository
	searchRepo *postgres.SearchRepository
	s3Storage  *s3storage.Storage
}

func NewDocumentService(
	docRepo *postgres.DocumentRepository,
	searchRepo *postgres.SearchRepository,
	s3Storage *s3storage.Storage,
) *DocumentService {
	return &DocumentService{
		docRepo:    docRepo,
		searchRepo: searchRepo,
		s3Storage:  s3Storage,
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

	contentType := mime.TypeByExtension(
		filepath.Ext(input.Filename),
	)

	if contentType == "" {
		contentType = "text/plain"
	}

	err := s.s3Storage.UploadFile(
		ctx,
		s3Key,
		input.Data,
		contentType,
	)
	if err != nil {
		return nil, err
	}

	title := input.Title
	if title == "" {
		title = input.Filename
	}

	doc := &model.Document{
		ID:               docID,
		OwnerID:          input.OwnerID,
		Title:            title,
		OriginalFilename: input.Filename,
		S3Key:            s3Key,
		IsPublic:         input.IsPublic,
		FileSize:         int64(len(input.Data)),
		MimeType:         contentType,
	}

	err = s.docRepo.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	err = s.indexDocument(ctx, docID, string(input.Data))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) indexDocument(
	ctx context.Context,
	docID uuid.UUID,
	text string,
) error {

	indexedTerms := indexer.BuildIndex(docID, text)

	totalDocs, err := s.searchRepo.CountDocuments(ctx)
	if err != nil {
		return err
	}

	if totalDocs == 0 {
		totalDocs = 1
	}

	for _, indexedTerm := range indexedTerms {

		tf := 1 + math.Log(float64(indexedTerm.Frequency))

		docsWithTerm, err := s.searchRepo.CountDocsWithTerm(ctx, indexedTerm.Term)
		if err != nil {
			return err
		}

		if docsWithTerm == 0 {
			continue
		}

		idf := math.Log(
			(float64(totalDocs) + 1) /
				(float64(docsWithTerm) + 1),
		)

		_ = tf * idf

		termID, err := s.searchRepo.CreateTerm(ctx, indexedTerm.Term)
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

		err = s.searchRepo.CreateTermPositions(
			ctx,
			postingID,
			indexedTerm.Positions,
		)
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
