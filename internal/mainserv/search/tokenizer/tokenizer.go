// Package tokenizer предоставляет надежную токенизацию с простым стеммингом
// для русского и английского языков.
//
// Принцип: один алгоритм для обоих языков, только разные списки суффиксов.
// Без словарей, без сложной логики.
//
// Не требует внешних зависимостей. Безопасен для UTF-8.
package tokenizer

import (
	"strings"
	"unicode"
)

var stopWords = map[string]struct{}{
	// Russian
	"в": {}, "во": {}, "на": {}, "по": {}, "о": {}, "об": {}, "обо": {},
	"с": {}, "со": {}, "из": {}, "изо": {}, "к": {}, "ко": {}, "у": {},
	"и": {}, "а": {}, "но": {}, "или": {}, "да": {}, "то": {}, "же": {},
	"ли": {}, "бы": {}, "не": {}, "ни": {}, "как": {}, "так": {}, "что": {},
	"это": {}, "тот": {}, "та": {}, "те": {}, "он": {}, "она": {}, "они": {},
	"за": {}, "над": {}, "под": {}, "от": {}, "до": {}, "без": {}, "для": {},
	"при": {}, "через": {}, "между": {}, "около": {}, "внутри": {},
	// English
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {},
	"by": {}, "for": {}, "from": {}, "has": {}, "he": {}, "in": {}, "is": {},
	"it": {}, "its": {}, "of": {}, "on": {}, "that": {}, "the": {}, "to": {},
	"was": {}, "will": {}, "with": {}, "this": {}, "but": {}, "they": {},
	"have": {}, "had": {}, "what": {}, "when": {}, "where": {}, "who": {},
	"which": {}, "why": {}, "how": {}, "all": {}, "each": {}, "every": {},
	"both": {}, "few": {}, "more": {}, "most": {}, "other": {}, "some": {},
	"such": {}, "no": {}, "nor": {}, "not": {}, "only": {}, "own": {},
	"same": {}, "so": {}, "than": {}, "too": {}, "very": {}, "can": {},
	"just": {}, "don": {}, "should": {}, "now": {},
}

func isVowel(r rune) bool {
	switch r {
	case 'а', 'е', 'ё', 'и', 'о', 'у', 'ы', 'э', 'ю', 'я',
		'a', 'e', 'i', 'o', 'u', 'y':
		return true
	}
	return false
}

func hasCyrillic(s string) bool {
	for _, r := range s {
		if r >= 'а' && r <= 'я' {
			return true
		}
	}
	return false
}

func hasVowel(s string) bool {
	for _, r := range s {
		if isVowel(r) {
			return true
		}
	}
	return false
}

var ruSuffixes = []string{
	"ирование", "ование", "ивание", "ывание", "яние", "ание", "ение",
	"ователь", "атель", "ятель", "тель", "ость", "еств", "стви", "ств",
	"аци", "изаци", "фика", "фикаци", "иров", "ова", "ева", "ива", "ыва",
	"ившись", "ывшись", "авшись", "явшись", "ись", "ась", "сь", "ясь",
	"авши", "ивши", "ывши", "явши", "вши", "авш", "ивш", "ывш", "явш",
	"ающ", "яющ", "ющ", "ющ", "аем", "яем", "им", "ым",
	"аешь", "яешь", "ешь", "ёшь", "ает", "яет", "ет", "ёт",
	"аем", "яем", "им", "ым", "аете", "яете", "ете", "ёте",
	"ают", "яют", "ут", "ют", "ат", "ят",
	"ила", "ыла", "ала", "яла", "ло", "ла", "ли", "л",
	"ится", "ытся", "ется", "ётся", "имся", "ымся", "емся", "ёмся",
	"итесь", "ытесь", "етесь", "ётесь",
	"ить", "ыть", "ать", "ять", "ть", "чь", "ти", "и", "ь", "й",
	"ующими", "яющими", "авшими", "ившими", "ывшими", "ющимися",
	"ующим", "яющим", "авшим", "ившим", "ывшим", "ющимся",
	"ующей", "яющей", "авшей", "ившей", "ывшей", "ющейся",
	"ующие", "яющие", "авшие", "ившие", "ывшие", "ящиеся",
	"ующего", "яющего", "авшего", "ившего", "ывшего", "ющегося",
	"аемый", "яемый", "имый", "енный", "анный", "янный",
	"ующ", "яющ", "авш", "ивш", "ывш", "явш",
	"ем", "аем", "яем", "им", "ым",
	"ете", "аете", "яете", "ются", "аются", "яются",
	"ется", "ается", "яется",
	"ешь", "аешь", "яешь",
	"ой", "ей", "ым", "им",
	"ого", "его", "ому", "ему",
	"ый", "ий",
	"ая", "яя", "ое", "ее",
	"иями", "ями", "ами",
	"иям", "ям",
	"иях", "ях", "ах", "ох", "ех", "их", "ых",
	"иев", "ьев", "ев", "ов",
	"ии", "ые", "ие", "ье", "е", "ия", "ья", "я", "ы", "и",
	"у", "ю", "о", "а", "ь", "ъ", "ом", "ем",
}

func stemRU(word string) string {
	if len([]rune(word)) < 4 || !hasCyrillic(word) {
		return word
	}

	result := word
	for pass := 0; pass < 5; pass++ {
		original := result
		for _, suf := range ruSuffixes {
			if strings.HasSuffix(result, suf) {
				candidate := strings.TrimSuffix(result, suf)
				if len([]rune(candidate)) >= 2 && hasVowel(candidate) {
					result = candidate
					break
				}
			}
		}
		if result == original {
			break
		}
	}
	return result
}

var enSuffixes = []string{
	// 7+ букв
	"ization", "isation", "ational", "tional", "ability", "ication",
	"iveness", "fulness", "ousness",
	// 6 букв
	"ation", "ition", "ution", "itive", "ative", "alize",
	"icity", "ivity", "fully", "ously",
	// 5 букв
	"ments", "ment", "ness", "able", "ible", "istic",
	// 4 буквы
	"ives", "ive", "izes", "ize", "ises", "ise", "ous",
	"ful", "ent", "ant", "ism", "ist", "ity", "ic", "al",
	"tion",
	// 3 буквы и меньше
	"ies", "es", "s", "ing", "ed", "er", "est", "e",
}

func stemEN(word string) string {
	if len([]rune(word)) < 4 {
		return strings.ToLower(word)
	}

	result := strings.ToLower(word)
	for pass := 0; pass < 3; pass++ {
		original := result
		for _, suf := range enSuffixes {
			if strings.HasSuffix(result, suf) {
				candidate := strings.TrimSuffix(result, suf)
				if len([]rune(candidate)) >= 3 && hasVowel(candidate) {
					result = candidate
					break
				}
			}
		}
		if result == original {
			break
		}
	}
	return result
}

func stem(word string) string {
	if hasCyrillic(word) {
		return stemRU(word)
	}
	return stemEN(word)
}

func Tokenize(text string) []string {
	text = strings.ToLower(text)
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return ' '
	}, text)

	words := strings.Fields(cleaned)
	seen := make(map[string]struct{}, len(words)/2)
	tokens := make([]string, 0, len(words)/2)

	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if _, isStop := stopWords[w]; isStop {
			continue
		}

		stemmed := stem(w)
		if len(stemmed) < 2 {
			continue
		}

		if _, exists := seen[stemmed]; !exists {
			seen[stemmed] = struct{}{}
			tokens = append(tokens, stemmed)
		}
	}

	return tokens
}
