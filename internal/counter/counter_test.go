package counter

import (
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"Empty string", "", 0},
		{"Whitespace only", "   ", 0},
		{"Six characters", "abcdef", 1},
		{"Seven characters", "abcdefg", 2},
		{"Twelve characters", "abcdefghijkl", 2},
		{"Thirteen characters", "abcdefghijklm", 3},
		{"With spaces", "hello world test", 3}, // 16 chars including spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountWords(tt.content)
			if result != tt.expected {
				t.Errorf("CountWords(%q) = %d, want %d", tt.content, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFolderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No spaces", "ProjectA", "ProjectA"},
		{"Single space", "Project A", "Project-A"},
		{"Multiple spaces", "My  Project  Name", "My-Project-Name"},
		{"Leading/trailing spaces", "  Project  ", "Project"},
		{"Mixed whitespace", "Project\tName\nTest", "Project-Name-Test"},
		{"Multiple hyphens", "Project---Name", "Project-Name"},
		{"Already has hyphens", "Project-Name", "Project-Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFolderName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFolderName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
