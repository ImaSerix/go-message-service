package message_test

import (
	"strings"
	"testing"

	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/google/uuid"
)

func TestContent(t *testing.T) {
	var longContent strings.Builder

	for i := 0; i < 1000; i++ {
		longContent.WriteString("0123456789")
	}
	longContent.WriteByte('0')
	tests := []struct {
		name    string
		content string
		expErr  error
	}{
		{name: "success", content: "nice content", expErr: nil},
		{name: "empty content", content: "", expErr: message.ErrContentEmpty},
		{name: "to long content", content: longContent.String(), expErr: message.ErrContentToLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := message.NewContent(test.content)
			if err != test.expErr {
				t.Fatalf("expected %v, got %v", test.expErr, err)
			}
			if err == nil && c != message.Content(test.content) {
				t.Fatalf("expected content %s, but got %s", test.content, c)
			}
			if err != nil && c != "" {
				t.Fatalf("expected empty content, but got %s", c)
			}
		})
	}
}

func TestNewMessage(t *testing.T) {

	id := message.SnowflakeID(1001)
	chatID := message.ChatID(uuid.New())
	senderID := message.SenderID(uuid.New())
	content, _ := message.NewContent("nice content")

	m := message.NewMessage(id, chatID, senderID, content)

	if m == nil {
		t.Fatal("message is nil")
	}
	if m.ID() != id {
		t.Fatal("id mismatch")
	}
	if m.ChatID() != chatID {
		t.Fatal("chatID mismatch")
	}
	if m.SenderID() != senderID {
		t.Fatal("sender mismatch")
	}
	if m.Content() != content {
		t.Fatal("content mismatch")
	}
}
