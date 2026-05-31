package chat

import (
	"strings"
)

type Name string

func NewName(name string) (Name, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) < 2 || len(trimmedName) > 32 {
		return "", ErrInvalidChatName
	}

	return Name(trimmedName), nil
}
