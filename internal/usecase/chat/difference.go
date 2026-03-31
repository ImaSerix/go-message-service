package chat

import "github.com/google/uuid"

func difference(a, b []uuid.UUID) []uuid.UUID {
	mb := make(map[uuid.UUID]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []uuid.UUID
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
