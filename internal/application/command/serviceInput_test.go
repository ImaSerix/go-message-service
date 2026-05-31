package command_test

import (
	"testing"

	"github.com/ImaSerix/go-message-service/internal/application/command"
	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/google/uuid"
)

func TestNewCreateChatInput(t *testing.T) {
	tests := []struct {
		name   string
		cmd    *command.CreateChatCmd
		expErr error
	}{
		{name: "success", cmd: &command.CreateChatCmd{CallerID: uuid.New(), Name: "nice name", Participants: []uuid.UUID{uuid.New(), uuid.New()}}, expErr: nil},
		{name: "cmd nil", cmd: nil, expErr: command.ErrInvalidCommand},
		{name: "not valid name", cmd: &command.CreateChatCmd{CallerID: uuid.New(), Name: "", Participants: []uuid.UUID{uuid.New(), uuid.New()}}, expErr: chat.ErrInvalidChatName},
	}

	for _, test := range tests {
		input, err := command.NewCreateChatInput(test.cmd)

		if err != test.expErr {
			t.Fatalf("expected %v error, got %v", test.expErr, err)
		}
		if err == nil && input == nil {
			t.Fatal("expected not nil input")
		}
		if err != nil && input != nil {
			t.Fatal("expected nil input on error")
		}
	}
}

func TestNewCreateMessageInput(t *testing.T) {
	tests := []struct {
		name   string
		cmd    *command.CreateMessageCmd
		expErr error
	}{
		{name: "success", cmd: &command.CreateMessageCmd{ChatID: uuid.New(), SenderID: uuid.New(), Content: "nice name"}, expErr: nil},
		{name: "cmd nil", cmd: nil, expErr: command.ErrInvalidCommand},
		{name: "not valid content", cmd: &command.CreateMessageCmd{ChatID: uuid.New(), SenderID: uuid.New(), Content: ""}, expErr: message.ErrContentEmpty},
	}

	for _, test := range tests {
		input, err := command.NewCreateMessageInput(test.cmd)

		if err != test.expErr {
			t.Fatalf("expected %v error, got %v", test.expErr, err)
		}
		if err == nil && input == nil {
			t.Fatal("expected not nil input")
		}
		if err != nil && input != nil {
			t.Fatal("expected nil input on error")
		}
	}
}

func TestNewReadMessageInput(t *testing.T) {
	tests := []struct {
		name   string
		cmd    *command.ReadMessageCmd
		expErr error
	}{
		{name: "success", cmd: &command.ReadMessageCmd{ChatID: uuid.New(), CallerID: uuid.New(), MessageID: 1001}, expErr: nil},
		{name: "cmd nil", cmd: nil, expErr: command.ErrInvalidCommand},
	}

	for _, test := range tests {
		input, err := command.NewReadMessageInput(test.cmd)

		if err != test.expErr {
			t.Fatalf("expected %v error, got %v", test.expErr, err)
		}
		if err == nil && input == nil {
			t.Fatal("expected not nil input")
		}
		if err != nil && input != nil {
			t.Fatal("expected nil input on error")
		}
	}
}

func TestNewCreateUserInput(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name   string
		cmd    *command.CreateUserCmd
		expErr error
	}{
		{name: "success", cmd: &command.CreateUserCmd{UserID: &userID, Username: "niceUsername"}, expErr: nil},
		{name: "no userID", cmd: &command.CreateUserCmd{UserID: nil, Username: "niceUsername"}, expErr: nil},
		{name: "cmd nil", cmd: nil, expErr: command.ErrInvalidCommand},
		{name: "not valid username", cmd: &command.CreateUserCmd{UserID: &userID, Username: ""}, expErr: user.ErrInvalidUsername},
	}

	for _, test := range tests {
		input, err := command.NewCreateUserInput(test.cmd)

		if err != test.expErr {
			t.Fatalf("expected %v error, got %v", test.expErr, err)
		}
		if err == nil && input == nil {
			t.Fatal("expected not nil input")
		}
		if err != nil && input != nil {
			t.Fatal("expected nil input on error")
		}
	}
}
