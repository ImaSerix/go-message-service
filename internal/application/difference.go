package application

import (
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/google/uuid"
)

func difference[T uuid.UUID | user.UserID](a, b []T) []T {
	mb := make(map[T]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []T
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
