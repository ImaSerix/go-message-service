package repository

import "errors"

var (
	ErrAlreadyExists     = errors.New("")
	ErrEmptyParticipants = errors.New("empty participants")
)
