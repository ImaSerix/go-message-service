package query

import "errors"

var (
	ErrMessageNotInChat = errors.New("message not in chat")
	ErrInvalidLimit     = errors.New("invalid limit")
)
