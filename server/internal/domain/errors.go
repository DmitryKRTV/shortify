package domain

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrPasswordTooShort = errors.New("password too short")
	ErrProfanity        = errors.New("profanity")
	ErrInvalidURL       = errors.New("invalid url")
	ErrInvalidCreds     = errors.New("invalid credentials")
	ErrForbidden        = errors.New("forbidden")
)
