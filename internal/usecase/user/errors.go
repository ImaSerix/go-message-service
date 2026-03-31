package user

import "errors"

var (
	ErrInvalidUUID     = errors.New("invalid UUID")
	ErrInvalidUsername = errors.New("invalid username")
	ErrAlreadyExists   = errors.New("already exists")
)
