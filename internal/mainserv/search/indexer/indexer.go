// Package indexer строит поисковый индекс
// на основе токенизированного текста.
package indexer

import (
	"github.com/google/uuid"

	"github.com/drobyshevv/doc-service/internal/mainserv/search/tokenizer"
)

// IndexedTerm описывает информацию о термине
// внутри конкретного документа.
type IndexedTerm struct {

	// Term — нормализованное слово.
	Term string

	// Frequency — количество вхождений слова в документе.
	Frequency int

	// Positions — позиции слова внутри документа.
	Positions []int

	// DocumentID — идентификатор документа.
	DocumentID uuid.UUID
}

// BuildIndex строит индекс документа.
//
// Для каждого уникального токена вычисляет:
//   - частоту встречаемости
//   - позиции в тексте
//
// Возвращает список IndexedTerm для сохранения в БД.
func BuildIndex(docID uuid.UUID, text string) []IndexedTerm {
	tokens := tokenizer.Tokenize(text)

	termMap := make(map[string]*IndexedTerm)

	for pos, token := range tokens {
		if _, exists := termMap[token]; !exists {
			termMap[token] = &IndexedTerm{
				Term:       token,
				Frequency:  0,
				Positions:  []int{},
				DocumentID: docID,
			}
		}

		termMap[token].Frequency++
		termMap[token].Positions = append(
			termMap[token].Positions,
			pos,
		)
	}

	result := make([]IndexedTerm, 0, len(termMap))

	for _, term := range termMap {
		result = append(result, *term)
	}

	return result
}
