package errors

import (
	"errors"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrCollectionEmpty    = errors.New("collection is empty")
)
