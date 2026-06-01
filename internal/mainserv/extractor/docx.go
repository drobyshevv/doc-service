package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

type DOCXExtractor struct{}

func (e *DOCXExtractor) Extract(r io.Reader, _ string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var docReader io.ReadCloser
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docReader, err = f.Open()
			if err != nil {
				return "", err
			}
			break
		}
	}
	if docReader == nil {
		return "", nil
	}
	defer docReader.Close()

	var text strings.Builder
	decoder := xml.NewDecoder(docReader)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "t" {
			var content string
			if err := decoder.DecodeElement(&content, &start); err == nil {
				text.WriteString(content)
				text.WriteString(" ")
			}
		}
	}
	return strings.TrimSpace(text.String()), nil
}
