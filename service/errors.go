package service

import "errors"

var (
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidOperation = errors.New("invalid operation")
	ErrAlreadyExists    = errors.New("already exists")
)
