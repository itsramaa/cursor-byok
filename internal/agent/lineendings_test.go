package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizeLineEndingsForFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "lineendings-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		existingFile   string // content of existing file, or "" if new file
		inputContent   string
		expectedOutput string
		description    string
	}{
		{
			name:           "new file with LF input on Windows",
			existingFile:   "",
			inputContent:   "line1\nline2\nline3",
			expectedOutput: "line1\r\nline2\r\nline3", // Windows default
			description:    "New files should use Windows CRLF by default",
		},
		{
			name:           "existing CRLF file with LF input",
			existingFile:   "line1\r\nline2\r\n",
			inputContent:   "line1\nline2\nline3",
			expectedOutput: "line1\r\nline2\r\nline3",
			description:    "Should preserve CRLF from existing file",
		},
		{
			name:           "existing LF file with CRLF input",
			existingFile:   "line1\nline2\n",
			inputContent:   "line1\r\nline2\r\nline3",
			expectedOutput: "line1\nline2\nline3",
			description:    "Should preserve LF from existing file",
		},
		{
			name:           "existing CRLF file with CRLF input",
			existingFile:   "line1\r\nline2\r\n",
			inputContent:   "line1\r\nline2\r\nline3",
			expectedOutput: "line1\r\nline2\r\nline3",
			description:    "Should preserve CRLF when both match",
		},
		{
			name:           "existing LF file with LF input",
			existingFile:   "line1\nline2\n",
			inputContent:   "line1\nline2\nline3",
			expectedOutput: "line1\nline2\nline3",
			description:    "Should preserve LF when both match",
		},
		{
			name:           "empty content",
			existingFile:   "line1\r\nline2\r\n",
			inputContent:   "",
			expectedOutput: "",
			description:    "Empty content should remain empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string

			if tt.existingFile != "" {
				// Create an existing file
				filePath = filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "_")+".txt")
				err := os.WriteFile(filePath, []byte(tt.existingFile), 0644)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				// New file path (doesn't exist yet)
				filePath = filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "_")+"_new.txt")
			}

			result := normalizeLineEndingsForFile(filePath, tt.inputContent)

			if result != tt.expectedOutput {
				t.Errorf("%s\nInput:    %q\nExpected: %q\nGot:      %q",
					tt.description, tt.inputContent, tt.expectedOutput, result)
			}
		})
	}
}

func TestDetectCRLF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "CRLF content",
			input:    "line1\r\nline2\r\n",
			expected: true,
		},
		{
			name:     "LF content",
			input:    "line1\nline2\n",
			expected: false,
		},
		{
			name:     "no line endings",
			input:    "single line",
			expected: runtime.GOOS == "windows", // defaults to OS convention
		},
		{
			name:     "empty string",
			input:    "",
			expected: runtime.GOOS == "windows", // defaults to OS convention
		},
		{
			name:     "mixed but starts with CRLF",
			input:    "line1\r\nline2\nline3",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectCRLF([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("detectCRLF(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
