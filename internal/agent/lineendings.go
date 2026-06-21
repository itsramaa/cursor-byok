package agent

import (
	"os"
	"runtime"
	"strings"
)

// normalizeLineEndingsForFile converts content to match the line-ending
// convention of the target file. Detection priority:
//  1. If the file already exists, sniff its first \r\n or bare \n to decide.
//  2. For new files, fall back to the OS default (\r\n on Windows, \n elsewhere).
//
// This is called on every Write / StrReplace payload before it's handed to
// Cursor's IDE via EditArgs.StreamContent. Without it, a model that emits
// \n for a Windows file (or vice versa) causes Cursor to write mixed line
// endings, which then shows up as double-spacing (\r\n\r\n) on subsequent
// reads and re-edits.
func normalizeLineEndingsForFile(path, content string) string {
	if content == "" {
		return content
	}
	// Step 1: always collapse to pure \n first so we have a clean baseline.
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Step 2: decide target convention.
	if targetCRLF(path) {
		return strings.ReplaceAll(content, "\n", "\r\n")
	}
	return content
}

// targetCRLF returns true when the file at path uses CRLF line endings,
// or when path is a new file and the OS default is CRLF (Windows).
func targetCRLF(path string) bool {
	// Try to detect from the existing file.
	if raw, err := os.ReadFile(path); err == nil && len(raw) > 0 {
		return detectCRLF(raw)
	}
	// New file — OS default.
	return runtime.GOOS == "windows"
}

// detectCRLF sniffs a byte slice for line-ending convention.
// Returns true when at least one \r\n is found before any bare \n.
func detectCRLF(data []byte) bool {
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			return true
		}
		if data[i] == '\n' {
			// Bare \n found first — file uses LF.
			return false
		}
	}
	// No newlines at all — default to OS convention.
	return runtime.GOOS == "windows"
}
