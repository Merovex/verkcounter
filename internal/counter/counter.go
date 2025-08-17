package counter

import (
	"regexp"
	"strings"
)

const CharactersPerWord = 6

type Result struct {
	FolderName string
	WordCount  int
}

func CountWords(content string) int {
	charCount := len(strings.TrimSpace(content))

	if charCount == 0 {
		return 0
	}

	wordCount := (charCount + CharactersPerWord - 1) / CharactersPerWord
	return wordCount
}

func SanitizeFolderName(name string) string {
	re := regexp.MustCompile(`\s+`)
	sanitized := re.ReplaceAllString(name, "-")

	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	sanitized = strings.Trim(sanitized, "-")

	return sanitized
}
