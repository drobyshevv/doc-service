// Package tokenizer предоставляет функции
// для разбиения текста на поисковые токены.
package tokenizer

import (
	"regexp"
	"strings"
)

// nonAlphaNumeric удаляет все символы,
// кроме букв и цифр.
var nonAlphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9а-яА-Я]+`)

// Tokenize преобразует текст в список токенов.
//
// Алгоритм:
//   - приводит текст к нижнему регистру
//   - удаляет спецсимволы
//   - разбивает строку по пробелам
//   - отбрасывает слишком короткие токены (< 2 символов)
func Tokenize(text string) []string {
	text = strings.ToLower(text)

	text = nonAlphaNumeric.ReplaceAllString(text, " ")

	rawTokens := strings.Fields(text)

	tokens := make([]string, 0, len(rawTokens))

	for _, token := range rawTokens {
		if len(token) < 2 {
			continue
		}

		tokens = append(tokens, token)
	}

	return tokens
}
