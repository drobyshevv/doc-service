package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/drobyshevv/doc-service/internal/mainserv/service"
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
func (h *DocumentHandler) UploadDocument(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	owner := r.FormValue("owner_id")
	ownerID, err := uuid.Parse(owner)
	if err != nil {
		http.Error(w, "invalid owner_id", http.StatusBadRequest)
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

	if ownerID != userID {
		http.Error(w, "forbidden: cannot upload for another user", http.StatusForbidden)
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
