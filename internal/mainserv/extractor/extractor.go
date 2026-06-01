package extractor

import (
	"io"
	"path/filepath"
	"strings"
)

// TextExtractor извлекает текстовое содержимое из документа.
type TextExtractor interface {
	Extract(r io.Reader, mimeType string) (string, error)
}

// NewExtractor возвращает подходящий экстрактор по MIME-типу или расширению.
func NewExtractor(mimeType string, filename string) TextExtractor {
	ext := strings.ToLower(filepath.Ext(filename))

	switch {
	case mimeType == "application/pdf" || ext == ".pdf":
		return &PDFExtractor{}
	case mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" || ext == ".docx":
		return &DOCXExtractor{}
	default:
		return &PlainTextExtractor{}
	}
}
