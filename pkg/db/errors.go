package db

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidID     = errors.New("invalid ID")
	ErrInternal      = errors.New("internal error")
)
