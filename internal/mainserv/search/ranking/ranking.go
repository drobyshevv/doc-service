// Package ranking отвечает за ранжирование
// результатов поиска.
package ranking

import (
	"math"
	"sort"
)

// SearchResult представляет документ,
// найденный по поисковому запросу.
type SearchResult struct {
	// DocumentID — идентификатор документа.
	DocumentID string

	// Score — итоговая релевантность документа.
	Score float64

	// DocLength — длина документа в токенах.
	DocLength int
}

// NormalizeScore нормализует score
// относительно длины документа.
//
// Более длинные документы получают меньший штраф,
// чтобы уменьшить bias в их пользу.
func NormalizeScore(score float64, docLength int) float64 {
	if docLength == 0 {
		return score
	}

	return score / math.Sqrt(float64(docLength))
}

// Rank нормализует score документов
// и сортирует результаты по релевантности
// в порядке убывания.
func Rank(results []SearchResult) []SearchResult {
	for i := range results {
		results[i].Score = NormalizeScore(
			results[i].Score,
			results[i].DocLength,
		)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
