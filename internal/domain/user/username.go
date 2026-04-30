package user

import "strings"

type Username string

func NewUsername(username string) (Username, error) {
	trimmedUsername := strings.TrimSpace(username)
	if trimmedUsername == "" || len(trimmedUsername) < 2 || len(trimmedUsername) > 32 {
		return "", ErrInvalidUsername
	}

	return Username(trimmedUsername), nil
}
