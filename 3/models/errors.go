package models

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("unexpected error")
)
