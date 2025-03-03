package file

import (
	"path/filepath"
	"strings"
)

func IsValidateFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".pdf", ".doc", ".docx":
		return true
	default:
		return false
	}
}
