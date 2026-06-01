package extractor

import (
	"bytes"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PDFExtractor struct{}

func (e *PDFExtractor) Extract(r io.Reader, _ string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	f, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var text strings.Builder
	pages := f.NumPage()
	for i := 1; i <= pages; i++ {
		page := f.Page(i)
		if page.V.IsNull() {
			continue
		}
		t, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text.WriteString(t)
		text.WriteString("\n")
	}
	return strings.TrimSpace(text.String()), nil
}
