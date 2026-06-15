package service

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed profanity_words.json
var profanityWordsJSON []byte

var (
	profanityWords     map[string]struct{}
	loadProfanityWords sync.Once
)

func profanityWordSet() map[string]struct{} {
	loadProfanityWords.Do(func() {
		var words []string
		if err := json.Unmarshal(profanityWordsJSON, &words); err != nil {
			profanityWords = map[string]struct{}{}
			return
		}

		profanityWords = make(map[string]struct{}, len(words))
		for _, word := range words {
			profanityWords[strings.ToLower(word)] = struct{}{}
		}
	})

	return profanityWords
}

func sanitizeProfanityInput(value string) string {
	replacer := strings.NewReplacer(".", " ", ",", " ")
	return strings.ToLower(strings.TrimSpace(replacer.Replace(value)))
}

func expandProfanityInput(value string) string {
	replacer := strings.NewReplacer(
		".", " ",
		"/", " ",
		":", " ",
		"?", " ",
		"&", " ",
		"=", " ",
		"_", " ",
		"-", " ",
	)
	return strings.ToLower(strings.TrimSpace(replacer.Replace(value)))
}

func textHasProfanityWord(text string) bool {
	words := profanityWordSet()
	if len(words) == 0 {
		return false
	}

	for _, token := range strings.Fields(text) {
		if _, ok := words[token]; ok {
			return true
		}
	}

	return false
}

func containsProfanity(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	return textHasProfanityWord(sanitizeProfanityInput(value)) ||
		textHasProfanityWord(expandProfanityInput(value))
}

func authContainsProfanity(email, password string) bool {
	return containsProfanity(email) || containsProfanity(password)
}
