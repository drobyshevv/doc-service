package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/drobyshevv/doc-service/internal/mainserv/model"
	"github.com/drobyshevv/doc-service/internal/mainserv/service"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage"
)

type DocumentHandler struct {
	documentService *service.DocumentService
}

func NewDocumentHandler(
	documentService *service.DocumentService,
) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// UploadDocument обрабатывает HTTP-запрос
// на загрузку нового документа.
//
// Ожидает multipart/form-data с файлом,
// метаданными документа и параметрами доступа.
// После успешной загрузки сохраняет файл,
// индексирует содержимое и возвращает созданный документ.
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "unauthorized: X-User-ID header required", http.StatusUnauthorized)
		return
	}

	ownerID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user header", http.StatusInternalServerError)
		return
	}

	input := service.UploadDocumentInput{
		OwnerID:  ownerID,
		Title:    r.FormValue("title"),
		Filename: fileHeader.Filename,
		Data:     data,
		IsPublic: r.FormValue("is_public") == "true",
	}

	doc, err := h.documentService.UploadDocument(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

// GetDocument обрабатывает HTTP-запрос
// на получение документа по его идентификатору.
//
// Возвращает содержимое файла и устанавливает
// соответствующие HTTP-заголовки для скачивания.
func (h *DocumentHandler) GetDocument(
	w http.ResponseWriter,
	r *http.Request,
) {
	idParam := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	doc, data, err := h.documentService.GetDocument(r.Context(), docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !doc.IsPublic {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "invalid user header", http.StatusInternalServerError)
			return
		}

		if doc.OwnerID != userID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+doc.OriginalFilename)
	w.Header().Set("Content-Type", doc.MimeType)
	w.Write(data)
}

// DeleteDocument обрабатывает HTTP-запрос
// на удаление документа.
//
// Удаляет файл из объектного хранилища,
// а также связанные метаданные и поисковый индекс.
func (h *DocumentHandler) DeleteDocument(
	w http.ResponseWriter,
	r *http.Request,
) {
	idParam := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	doc, _, err := h.documentService.GetDocument(r.Context(), docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user header", http.StatusInternalServerError)
		return
	}

	if doc.OwnerID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	err = h.documentService.DeleteDocument(r.Context(), docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadNewVersion загружает новую версию документа.
func (h *DocumentHandler) UploadNewVersion(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, _ := io.ReadAll(file)
	note := r.FormValue("note") // опциональный комментарий

	version, err := h.documentService.UploadNewVersion(
		r.Context(), docID, userID, data, header.Filename, note,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// ListVersions получает список версий документа.
func (h *DocumentHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	versions, err := h.documentService.ListDocumentVersions(r.Context(), docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

// GetVersion скачивает конкретную версию документа.
func (h *DocumentHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	versionNum, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		http.Error(w, "invalid version", http.StatusBadRequest)
		return
	}

	ver, data, err := h.documentService.GetDocumentVersion(r.Context(), docID, versionNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=v%d-%s", ver.Version, ver.S3Key))
	w.Header().Set("Content-Type", ver.MimeType)
	w.Write(data)
}

// RollbackVersion откатывает к предыдущей версии документа.
func (h *DocumentHandler) RollbackVersion(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	versionNum, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		http.Error(w, "invalid version", http.StatusBadRequest)
		return
	}

	err = h.documentService.RollbackToVersion(r.Context(), docID, versionNum, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMetadata обрабатывает запрос на обновление метаданных документа.
func (h *DocumentHandler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input model.UpdateMetadataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updatedDoc, err := h.documentService.UpdateDocumentMetadata(
		r.Context(), docID, userID, input,
	)
	if err != nil {
		if err.Error() == "forbidden: not document owner" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedDoc)
}

// ListDocuments возвращает ТОЛЬКО документы текущего пользователя.
// Вызывается со страницы "Мои документы" (/).
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user header", http.StatusInternalServerError)
		return
	}

	documents, err := h.documentService.ListByOwner(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": documents,
	})
}

// ListPublicDocuments возвращает ТОЛЬКО публичные документы всех пользователей.
// Вызывается со страницы поиска публичных (/search).
func (h *DocumentHandler) ListPublicDocuments(w http.ResponseWriter, r *http.Request) {

	documents, _, err := h.documentService.ListDocuments(r.Context(), uuid.Nil, model.ListDocumentsQuery{
		Limit:    100,
		IsPublic: boolPtr(true),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": documents,
	})
}

func boolPtr(b bool) *bool {
	return &b
}

// GetMetadata обрабатывает запрос на получение метаданных документа без скачивания файла.
func (h *DocumentHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var requesterID uuid.UUID
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		requesterID, _ = uuid.Parse(uid)
	}

	meta, err := h.documentService.GetDocumentMetadata(r.Context(), docID, requesterID)
	if err != nil {
		if err.Error() == "forbidden" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err == storage.ErrDocumentNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}
