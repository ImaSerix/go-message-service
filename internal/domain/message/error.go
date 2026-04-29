package message

import "errors"

var (
	ErrContentEmpty  = errors.New("message content empty")
	ErrContentToLong = errors.New("content to long")
)
