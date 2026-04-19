package chat_test

import (
	"testing"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
)

func TestName(t *testing.T) {
	tests := []struct {
		name        string
		chatName    string
		expChatName string
		expErr      error
	}{
		{name: "success", chatName: "niceName", expChatName: "niceName", expErr: nil},
		{name: "short name", chatName: "ab", expChatName: "ab", expErr: nil},
		{name: "long name", chatName: "longlonglonglonglonglonglongname", expChatName: "longlonglonglonglonglonglongname", expErr: nil},
		{name: "long name with spaces", chatName: "     longlonglonglonglonglonglongname     ", expChatName: "longlonglonglonglonglonglongname", expErr: nil},
		{name: "with spaces name", chatName: " name    ", expChatName: "name", expErr: nil},

		{name: "empty name", chatName: "", expErr: chat.ErrInvalidChatName},
		{name: "very short name", chatName: "a", expErr: chat.ErrInvalidChatName},
		{name: "very long name", chatName: "vlonglonglonglonglonglonglongname", expErr: chat.ErrInvalidChatName},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			n, err := chat.NewName(test.chatName)
			if err != test.expErr {
				t.Fatalf("expected %v, got %v", test.expErr, err)
			}
			if string(n) != test.expChatName {
				t.Fatalf("expected %s, got %s", test.expChatName, n)
			}
		})
	}
}
