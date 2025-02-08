package db

import "errors"

var (
	ErrNotFound      = errors.New("Not found")
	ErrAlreadyExists = errors.New("Already Exists")
	ErrInvalidName   = errors.New("Invalid Name")
	ErrInternal      = errors.New("Internal Error")
)
