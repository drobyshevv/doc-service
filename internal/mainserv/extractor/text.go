package extractor

import (
	"io"
	"strings"
)

type PlainTextExtractor struct{}

func (e *PlainTextExtractor) Extract(r io.Reader, _ string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
