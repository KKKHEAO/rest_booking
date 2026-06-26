package domain

import "errors"

// Доменные ошибки.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
	ErrTableTaken = errors.New("table already booked")
)
