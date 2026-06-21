package proximity

import (
	"math"
	"sort"
)

type positionWithTerm struct {
	pos     int
	termIdx int
}

// CalculateProximityBonus рассчитывает бонус за близость слов в документе.
// positions: map[termIdx][]int - позиции каждого термина запроса в документе.
// Возвращает значение от 0.0 (слова далеко) до 1.0 (слова строго подряд).
func CalculateProximityBonus(positions map[int][]int) float64 {
	if len(positions) <= 1 {
		return 0
	}
	var allPositions []positionWithTerm
	for termIdx, posList := range positions {
		if len(posList) > 20 {
			posList = posList[:20]
		}
		for _, pos := range posList {
			allPositions = append(allPositions, positionWithTerm{
				pos:     pos,
				termIdx: termIdx,
			})
		}
	}

	if len(allPositions) == 0 {
		return 0
	}

	sort.Slice(allPositions, func(i, j int) bool {
		return allPositions[i].pos < allPositions[j].pos
	})

	totalTerms := len(positions)
	termCount := make(map[int]int)
	uniqueInWindow := 0
	minWindow := math.MaxInt32
	left := 0
	for right := 0; right < len(allPositions); right++ {
		rightTerm := allPositions[right].termIdx
		if termCount[rightTerm] == 0 {
			uniqueInWindow++
		}
		termCount[rightTerm]++

		for uniqueInWindow == totalTerms {
			window := allPositions[right].pos - allPositions[left].pos + 1
			if window < minWindow {
				minWindow = window
			}

			leftTerm := allPositions[left].termIdx
			termCount[leftTerm]--
			if termCount[leftTerm] == 0 {
				uniqueInWindow--
			}
			left++
		}
	}

	if minWindow == math.MaxInt32 {
		return 0
	}
	return float64(totalTerms) / float64(minWindow)
}
