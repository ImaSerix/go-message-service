package user_test

import (
	"testing"

	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/google/uuid"
)

func TestUsername(t *testing.T) {

	tests := []struct {
		name        string
		id          uuid.UUID
		username    string
		expUsername string
		expErr      error
	}{
		{name: "success", username: "niceUsername", expUsername: "niceUsername", expErr: nil},
		{name: "short username", username: "ab", expUsername: "ab", expErr: nil},
		{name: "long username", username: "longlonglonglonglonglongusername", expUsername: "longlonglonglonglonglongusername", expErr: nil},
		{name: "long username with spaces", username: "     longlonglonglonglonglongusername     ", expUsername: "longlonglonglonglonglongusername", expErr: nil},
		{name: "with spaces username", username: " username    ", expUsername: "username", expErr: nil},

		{name: "empty username", username: "", expErr: user.ErrInvalidUsername},
		{name: "very short username", username: "a", expErr: user.ErrInvalidUsername},
		{name: "very long username", username: "vlonglonglonglonglonglongusername", expErr: user.ErrInvalidUsername},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u, err := user.NewUsername(test.username)
			if err != test.expErr {
				t.Fatalf("expected %v, got %v", test.expErr, err)
			}
			if string(u) != test.expUsername {
				t.Fatalf("expected %s, got %s", test.expUsername, u)
			}
		})
	}
}

func TestUser(t *testing.T) {

	id := user.UserID(uuid.New())
	username, _ := user.NewUsername("test")

	u := user.NewUser(id, username)

	if u == nil {
		t.Fatal("user is nil")
	}
	if u.ID() != id {
		t.Fatal("id mismatch")
	}
	if u.Username() != username {
		t.Fatal("username mismatch")
	}
}
