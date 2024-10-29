package datastore

import (
	"errors"
)

// ErrDBDocumentNotFound is returned when a document is not found in the database.
var ErrDBDocumentNotFound = errors.New("document not found in database")

var ErrDBDatasetExists = errors.New("dataset already exists in database")
