package types

import (
	"errors"
)

// ErrDBDocumentNotFound is returned when a document is not found in the database.
var ErrDBDocumentNotFound = errors.New("document not found in database")

var ErrDBDatasetExists = errors.New("dataset already exists in database")

// ErrDBFileNotFound is returned when a file is not found.
var ErrDBFileNotFound = errors.New("file not found in database")
