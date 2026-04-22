package message_test

import (
	"testing"

	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/google/uuid"
)

// TODO: фактически не известно, что будет, если content слишком большой, по логике лучше щто ограничить, но наверное хватит на данный момент просто упоменуть это в readme, как уязвимость
func TestContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		expErr  error
	}{
		{name: "success", content: "nice content", expErr: nil},
		{name: "empty content", content: "", expErr: message.ErrContentEmpty},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := message.NewContent(test.content)
			if err != test.expErr {
				t.Fatalf("expected %v, got %v", test.expErr, err)
			}
			if c != message.Content(test.content) {
				t.Fatal("content mismatch")
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
