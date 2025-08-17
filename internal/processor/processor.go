package processor

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ProcessMarkdownFiles(folderPath string) (string, error) {
	var allContent strings.Builder

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			content, err := processMarkdownFile(path)
			if err != nil {
				return nil
			}
			allContent.WriteString(content)
			allContent.WriteString(" ")
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return allContent.String(), nil
}

func processMarkdownFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return stripFrontmatter(string(content)), nil
}

func stripFrontmatter(content string) string {
	lines := strings.Split(content, "\n")

	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		inFrontmatter := true
		startIdx := 1

		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				inFrontmatter = false
				startIdx = i + 1
				break
			}
		}

		if !inFrontmatter && startIdx < len(lines) {
			return strings.Join(lines[startIdx:], "\n")
		}
	}

	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "---") {
		parts := bytes.Split([]byte(content), []byte("---"))
		if len(parts) >= 3 {
			remainingParts := parts[2:]
			result := bytes.Join(remainingParts, []byte("---"))
			return string(result)
		}
	}

	return content
}
