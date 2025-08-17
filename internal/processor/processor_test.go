package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "No frontmatter",
			content:  "# Hello World\n\nThis is content.",
			expected: "# Hello World\n\nThis is content.",
		},
		{
			name: "With frontmatter",
			content: `---
title: Test
author: Someone
---

# Hello World

This is content.`,
			expected: "\n# Hello World\n\nThis is content.",
		},
		{
			name: "Frontmatter only",
			content: `---
title: Test
---`,
			expected: "",
		},
		{
			name:     "Invalid frontmatter",
			content:  "---\nNot closed properly\n\n# Content",
			expected: "---\nNot closed properly\n\n# Content",
		},
		{
			name: "Multiple dashes in content",
			content: `---
title: Test
---

# Content

Some text --- more text`,
			expected: "\n# Content\n\nSome text --- more text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripFrontmatter(tt.content)
			if result != tt.expected {
				t.Errorf("stripFrontmatter() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessMarkdownFiles(t *testing.T) {
	tempDir := t.TempDir()

	testFile1 := filepath.Join(tempDir, "test1.md")
	content1 := `---
title: Test1
---

# Test Document 1

This is test content.`

	err := os.WriteFile(testFile1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testFile2 := filepath.Join(tempDir, "test2.md")
	content2 := `# Test Document 2

No frontmatter here.`

	err = os.WriteFile(testFile2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	nonMdFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(nonMdFile, []byte("Not a markdown file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ProcessMarkdownFiles(tempDir)
	if err != nil {
		t.Fatalf("ProcessMarkdownFiles failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if len(result) == 0 {
		t.Error("Result should contain processed content")
	}
}
