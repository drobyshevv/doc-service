// Package tfidf содержит функции
// для вычисления TF-IDF метрик.
package tfidf

import "math"

// CalculateTF вычисляет term frequency.
//
// Формула:
//
//	TF = termCount / totalTerms
func CalculateTF(termCount int, totalTerms int) float64 {
	if totalTerms == 0 {
		return 0
	}

	return float64(termCount) / float64(totalTerms)
}

// CalculateIDF вычисляет inverse document frequency.
//
// Формула:
//
//	IDF = log(totalDocs / docsWithTerm)
func CalculateIDF(totalDocs int, docsWithTerm int) float64 {
	if docsWithTerm == 0 {
		return 0
	}

	return math.Log(float64(totalDocs) / float64(docsWithTerm))
}

// CalculateTFIDF вычисляет итоговый TF-IDF score.
//
// Формула:
//
//	TF-IDF = TF * IDF
func CalculateTFIDF(tf, idf float64) float64 {
	return tf * idf
}
