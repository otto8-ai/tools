package filetypes

import (
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

var FirstclassFileExtensions = map[string]struct{}{
	".pdf":   {},
	".html":  {},
	".md":    {},
	".txt":   {},
	".docx":  {},
	".odt":   {},
	".rtf":   {},
	".csv":   {},
	".ipynb": {},
	".json":  {},
}

// GetFiletype returns the filetype of a file based on its filename or content.
func GetFiletype(filename string, content []byte) (string, error) {
	// 1. By file extension, if available and first-class supported
	ext := path.Ext(filename)
	if _, ok := FirstclassFileExtensions[ext]; ok {
		return ext, nil
	}

	// 2. By content (mimetype)
	mt := mimetype.Detect(content)
	if mt != nil {
		return strings.Split(mt.String(), ";")[0], nil // remove charset (mimetype), e.g. from "text/plain; charset=utf-8"
	}

	slog.Error("Failed to detect filetype", "filename", filename)
	return "", fmt.Errorf("failed to detect filetype")
}
