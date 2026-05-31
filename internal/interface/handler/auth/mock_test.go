package auth_test

import (
	"context"

	"github.com/google/uuid"
)

type fakeUsersQry struct {
	called bool
	exists bool
	err    error
}

func (q *fakeUsersQry) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	q.called = true
	return q.exists, q.err
}
