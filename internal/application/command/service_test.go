package command_test

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ImaSerix/go-message-service/internal/application/command"
	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

func TestHandleCreateChat_InvalidInput(t *testing.T) {

	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	var cmd *command.CreateChatCmd
	newChatID, err := cmdServ.HandleCreateChat(t.Context(), cmd)
	if !errors.Is(err, command.ErrInvalidCommand) {
		t.Fatalf("expected wrapped error %v, got %v", command.ErrInvalidCommand, err)
	}
	if newChatID != uuid.Nil {
		t.Fatalf("expected uuid %s (nil), got %s", uuid.Nil, newChatID)
	}
	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected ID generator NOT to be called, but it was")
	}
}

func TestHandleCreateChat_DomainError(t *testing.T) {

	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := command.CreateChatCmd{
		CallerID:     uuid.New(),
		Name:         "nice name",
		Participants: []uuid.UUID{uuid.New(), uuid.New()},
	}

	newChatID, err := cmdServ.HandleCreateChat(t.Context(), &cmd)
	if !errors.Is(err, chat.ErrCallerIsNotParticipant) {
		t.Fatalf("expected wrapped error %v, got %v", chat.ErrCallerIsNotParticipant, err)
	}
	if newChatID != uuid.Nil {
		t.Fatalf("expected uuid %s (nil), got %s", uuid.Nil, newChatID)
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected ID generator NOT to be called, but it was")
	}
}

func TestHandleCreateChat_MissingUser(t *testing.T) {
	callerID := uuid.New()
	participants := []uuid.UUID{callerID, uuid.New(), uuid.New(), uuid.New()}
	tests := []struct {
		name       string
		cmd        *command.CreateChatCmd
		exists     []user.UserID
		expMissing []user.UserID
	}{
		{
			name: "one missing",
			cmd: &command.CreateChatCmd{
				CallerID:     callerID,
				Name:         "nice name",
				Participants: participants,
			},
			exists:     []user.UserID{user.UserID(participants[0]), user.UserID(participants[1]), user.UserID(participants[2])},
			expMissing: []user.UserID{user.UserID(participants[3])}},
		{
			name: "all missing",
			cmd: &command.CreateChatCmd{
				CallerID:     callerID,
				Name:         "nice name",
				Participants: participants,
			},
			exists:     []user.UserID{},
			expMissing: []user.UserID{user.UserID(participants[0]), user.UserID(participants[1]), user.UserID(participants[2]), user.UserID(participants[3])},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()
			fUsers.existingIDs = test.exists

			newChatID, err := cmdServ.HandleCreateChat(t.Context(), test.cmd)
			if err == nil {
				t.Fatal("expected non-nil error, but got nil")
			}
			typedError, ok := err.(*command.ErrUsersNotExist)
			if !ok {
				t.Fatal("expected error UserNotExists, but got something other")
			}
			if !slices.Equal(typedError.Users, test.expMissing) {
				t.Fatalf("expected users missing %v, but got %v", test.expMissing, typedError.Users)
			}
			if newChatID != uuid.Nil {
				t.Fatalf("expected uuid %s (nil), got %s", uuid.Nil, newChatID)
			}
			if !fUsers.called {
				t.Fatal("expected users to be called, but it was not")
			}
			if fChats.called {
				t.Fatal("expected chats NOT to be called, but it was")
			}
			if fNotifier.called {
				t.Fatal("expected notifier NOT to be called, but it was")
			}
			if fMessages.called {
				t.Fatal("expected messages NOT to be called, but it was")
			}
			if fGenerator.called {
				t.Fatal("expected ID generator NOT to be called, but it was")
			}
		})
	}
}

func TestHandleCreateChat_RepoError(t *testing.T) {

	callerID := uuid.New()
	cmd := &command.CreateChatCmd{
		CallerID:     callerID,
		Name:         "nice name",
		Participants: []uuid.UUID{callerID, uuid.New()},
	}

	tests := func(
		t *testing.T,
		cmdServ *command.CommandService,
		fMessages *fakeMessagesRepo,
		fUsers *fakeUsersRepo,
		fNotifier *fakeNotifier,
		fGenerator *fakeIDGenerator,
	) {
		fUsers.existingIDs = []user.UserID{user.UserID(cmd.Participants[0]), user.UserID(cmd.Participants[1])}
		newChatID, err := cmdServ.HandleCreateChat(t.Context(), cmd)

		if !strings.Contains(err.Error(), "create chat") {
			t.Fatalf("expected wrapped error containing 'create chat', but got %v", err)
		}
		if newChatID != uuid.Nil {
			t.Fatalf("expected uuid %s (nil), but got %s", uuid.Nil, newChatID)
		}
		if !fUsers.called {
			t.Fatal("expected users NOT to be called, but it was")
		}
		if fNotifier.called {
			t.Fatal("expected notifier NOT to be called, but it was")
		}
		if fMessages.called {
			t.Fatal("expected messages NOT to be called, but it was")
		}
		if fGenerator.called {
			t.Fatal("expected ID generator NOT to be called, but it was")
		}
	}

	t.Run("user repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fUsers.err = errors.New("repo error")

		tests(t, cmdServ, fMessages, fUsers, fNotifier, fGenerator)
		if fChats.called {
			t.Fatal("expected chats NOT to be called, but it was")
		}
	})

	t.Run("chat repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fChats.err = errors.New("repo error")

		tests(t, cmdServ, fMessages, fUsers, fNotifier, fGenerator)
		if !fChats.called {
			t.Fatal("expected chats to be called, but it was not")
		}
	})

}

func TestHandleCreateChat_AlreadyExists(t *testing.T) {
	callerID := uuid.New()
	participants := []uuid.UUID{callerID, uuid.New()}
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.CreateChatCmd{
		CallerID:     callerID,
		Name:         "nice name",
		Participants: participants,
	}

	fChats.err = command.ErrAlreadyExists
	fUsers.existingIDs = []user.UserID{user.UserID(participants[0]), user.UserID(participants[1])}

	newChatID, err := cmdServ.HandleCreateChat(t.Context(), cmd)
	if !errors.Is(err, command.ErrAlreadyExists) {
		t.Fatalf("expected wrapped error %v, but got %v", command.ErrAlreadyExists, err)
	}
	if newChatID != uuid.Nil {
		t.Fatalf("expected uuid %s (nil), got %s", uuid.Nil, newChatID)
	}
	if !fUsers.called {
		t.Fatal("expected users to be called, but it was not")
	}
	if !fChats.called {
		t.Fatal("expected chats to be called, but it was not")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected ID generator NOT to be called, but it was")
	}
}

func TestHandleCreateChat_Success(t *testing.T) {
	callerID := uuid.New()
	participants := []uuid.UUID{callerID, uuid.New()}
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.CreateChatCmd{
		CallerID:     callerID,
		Name:         "nice name",
		Participants: participants,
	}

	fChats.chat = chat.RehydrateChat(
		chat.ChatID(uuid.New()),
		chat.Name(cmd.Name),
		[]chat.ParticipantID{chat.ParticipantID(participants[0]), chat.ParticipantID(participants[1])},
		nil,
		nil,
	)
	fUsers.existingIDs = []user.UserID{user.UserID(participants[0]), user.UserID(participants[1])}

	newChatID, err := cmdServ.HandleCreateChat(t.Context(), cmd)
	if err != nil {
		t.Fatalf("expected nil error, but got %v", err)
	}
	if newChatID != uuid.UUID(fChats.chat.ID()) {
		t.Fatalf("expected uuid %s, but got %s", fChats.chat.ID(), newChatID)
	}
	if !fUsers.called {
		t.Fatal("expected users to be called, but it was not")
	}
	if !fChats.called {
		t.Fatal("expected chats to be called, but it was not")
	}
	if !fNotifier.called {
		t.Fatal("expected notifier to be called, but it was not")
	}
	typedEvent, ok := fNotifier.event.(*wsevent.NewChat)
	if !ok {
		t.Fatal("expected notifier to get NewChat event, but it was not")
	}
	if typedEvent.Chat.ID != newChatID {
		t.Fatalf("expected chat id %s, but got %s", newChatID, typedEvent.Chat.ID)
	}
	if typedEvent.Chat.Name != cmd.Name {
		t.Fatalf("expected chat name %s, but got %s", cmd.Name, typedEvent.Chat.Name)
	}
	if !slices.Equal(typedEvent.Chat.Participants, cmd.Participants) {
		t.Fatalf("expected participants %v, but got %v", cmd.Participants, typedEvent.Chat.Participants)
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected ID generator NOT to be called, but it was")
	}
}

func TestHandleCreateMessage_InvalidInput(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	var cmd *command.CreateMessageCmd
	newMessageID, err := cmdServ.HandleCreateMessage(t.Context(), cmd)
	if !errors.Is(err, command.ErrInvalidCommand) {
		t.Fatalf("expected wrapped error %v, but got %v", command.ErrInvalidCommand, err)
	}
	if newMessageID != 0 {
		t.Fatalf("expected zero id, but got %d", newMessageID)
	}
	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleCreateMessage_ChatNotExists(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()
	fChats.err = command.ErrNotFound

	cmd := &command.CreateMessageCmd{
		ChatID:   uuid.New(),
		SenderID: uuid.New(),
		Content:  "nice content",
	}

	newMessageID, err := cmdServ.HandleCreateMessage(t.Context(), cmd)
	if !errors.Is(err, command.ErrChatNotExists) {
		t.Fatalf("expected wrapped error %v, but got %v", command.ErrChatNotExists, err)
	}
	if newMessageID != 0 {
		t.Fatalf("expected zero id, but got %d", newMessageID)
	}
	if !fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleCreateMessage_DomainError(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()
	fChats.chat = chat.RehydrateChat(
		chat.ChatID(uuid.New()),
		"nice name",
		[]chat.ParticipantID{},
		nil,
		nil,
	)

	cmd := &command.CreateMessageCmd{
		ChatID:   uuid.New(),
		SenderID: uuid.New(),
		Content:  "nice content",
	}

	newMessageID, err := cmdServ.HandleCreateMessage(t.Context(), cmd)
	if !errors.Is(err, chat.ErrCallerIsNotParticipant) {
		t.Fatalf("expected wrapped error %v, but got %v", chat.ErrCallerIsNotParticipant, err)
	}
	if newMessageID != 0 {
		t.Fatalf("expected zero id, but got %d", newMessageID)
	}
	if !fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if !fGenerator.called {
		t.Fatal("expected id generator to be called, but it was not")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleCreateMessage_RepoError(t *testing.T) {

	cmd := &command.CreateMessageCmd{
		ChatID:   uuid.New(),
		SenderID: uuid.New(),
		Content:  "nice content",
	}

	tests := func(
		t *testing.T,
		cmdServ *command.CommandService,
		fChats *fakeChatsRepo,
		fUsers *fakeUsersRepo,
		fNotifier *fakeNotifier,
	) {
		newMessageID, err := cmdServ.HandleCreateMessage(t.Context(), cmd)

		if !strings.Contains(err.Error(), "handleCreateMessage") {
			t.Fatalf("expected wrapped error containing 'handleCreateMessage', but got %v", err)
		}
		if newMessageID != 0 {
			t.Fatalf("expected zero id, but got %d", newMessageID)
		}
		if !fChats.called {
			t.Fatal("expected chats NOT to be called, but it was")
		}
		if fNotifier.called {
			t.Fatal("expected notifier NOT to be called, but it was")
		}
		if fUsers.called {
			t.Fatal("expected users NOT to be called, but it was")
		}
	}

	t.Run("chat repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fChats.err = errors.New("repo error")

		tests(t, cmdServ, fChats, fUsers, fNotifier)
		if fGenerator.called {
			t.Fatal("expected id generator NOT to be called, but it was")
		}
		if fMessages.called {
			t.Fatal("expected messages NOT to be called, but it was")
		}
	})

	t.Run("message repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fMessages.err = errors.New("repo error")
		fChats.chat = chat.RehydrateChat(
			chat.ChatID(uuid.New()),
			"nice name",
			[]chat.ParticipantID{chat.ParticipantID(cmd.SenderID), chat.ParticipantID(uuid.New())},
			nil,
			nil,
		)

		tests(t, cmdServ, fChats, fUsers, fNotifier)
		if !fGenerator.called {
			t.Fatal("expected id generator to be called, but it was not")
		}
		if !fMessages.called {
			t.Fatal("expected messages to be called, but it was not")
		}
	})
}

func TestHandleCreateMessage_Success(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()
	cmd := &command.CreateMessageCmd{
		ChatID:   uuid.New(),
		SenderID: uuid.New(),
		Content:  "nice content",
	}

	fChats.chat = chat.RehydrateChat(
		chat.ChatID(uuid.New()),
		"nice name",
		[]chat.ParticipantID{chat.ParticipantID(cmd.SenderID), chat.ParticipantID(uuid.New())},
		nil,
		nil,
	)
	fMessages.message = message.NewMessage(
		message.SnowflakeID(1001),
		message.ChatID(cmd.ChatID),
		message.SenderID(cmd.SenderID),
		message.Content(cmd.Content),
	)
	fMessages.message.SetCreatedAt(time.Now())

	newMessageID, err := cmdServ.HandleCreateMessage(t.Context(), cmd)
	if err != nil {
		t.Fatalf("expected nil error, but got %v", err)
	}
	if newMessageID == 0 {
		t.Fatal("expected non-zero id, but got zero")
	}
	if !fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if !fGenerator.called {
		t.Fatal("expected id generator to be called, but it was not")
	}
	if !fMessages.called {
		t.Fatal("expected messages to be called, but it was not")
	}
	if !fNotifier.called {
		t.Fatal("expected notifier to be called, but it was not")
	}
	typedEvent, ok := fNotifier.event.(*wsevent.NewMessage)
	if !ok {
		t.Fatal("expected notifier to get NewMessage event, but it was not")
	}
	if typedEvent.Message.ID != newMessageID {
		t.Fatalf("expected message id %d, but got %d", newMessageID, typedEvent.Message.ID)
	}
	if typedEvent.Message.ChatID != cmd.ChatID {
		t.Fatalf("expected chat id %s, but got %s", cmd.ChatID, typedEvent.Message.ChatID)
	}
	if typedEvent.Message.Content != cmd.Content {
		t.Fatalf("expected content %s, but got %s", cmd.Content, typedEvent.Message.Content)
	}
	if typedEvent.Message.SenderID != cmd.SenderID {
		t.Fatalf("expected sender id %s, but got %s", cmd.SenderID, typedEvent.Message.SenderID)
	}

	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleMarkAsRead_InvalidInput(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	var cmd *command.ReadMessageCmd
	err := cmdServ.HandleMarkAsRead(t.Context(), cmd)
	if !errors.Is(err, command.ErrInvalidCommand) {
		t.Fatalf("expected wrapped error %v, but got %v", command.ErrInvalidCommand, err)
	}
	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}

	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleMarkAsRead_ChatNotExists(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.ReadMessageCmd{
		ChatID:    uuid.New(),
		CallerID:  uuid.New(),
		MessageID: 1001,
	}
	fChats.err = command.ErrNotFound
	err := cmdServ.HandleMarkAsRead(t.Context(), cmd)
	if !errors.Is(err, command.ErrChatNotExists) {
		t.Fatalf("expected wrapped error %v, but got %v", command.ErrChatNotExists, err)
	}
	if !fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}

	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleMarkAsRead_DomainError(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.ReadMessageCmd{
		ChatID:    uuid.New(),
		CallerID:  uuid.New(),
		MessageID: 1001,
	}
	fChats.chat = chat.RehydrateChat(
		chat.ChatID(cmd.ChatID),
		chat.Name("nice name"),
		[]chat.ParticipantID{
			chat.ParticipantID(cmd.CallerID),
			chat.ParticipantID(uuid.New()),
		},
		nil,
		nil,
	)
	fMessages.existsInChat = false

	err := cmdServ.HandleMarkAsRead(t.Context(), cmd)
	if !errors.Is(err, chat.ErrMessageNotInChat) {
		t.Fatalf("expected wrapped error %v, but got %v", chat.ErrMessageNotInChat, err)
	}
	if !fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if !fMessages.called {
		t.Fatal("expected messages to be called, but it was not")
	}

	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleMarkAsRead_RepoError(t *testing.T) {
	cmd := &command.ReadMessageCmd{
		ChatID:    uuid.New(),
		CallerID:  uuid.New(),
		MessageID: 1001,
	}

	tests := func(
		t *testing.T,
		cmdServ *command.CommandService,
		fChats *fakeChatsRepo,
		fUsers *fakeUsersRepo,
		fNotifier *fakeNotifier,
		fGenerator *fakeIDGenerator,
	) {
		err := cmdServ.HandleMarkAsRead(t.Context(), cmd)

		if !strings.Contains(err.Error(), "markAsRead") {
			t.Fatalf("expected wrapped error containing 'markAsRead', but got %v", err)
		}

		if !fChats.called {
			t.Fatal("expected chats NOT to be called, but it was")
		}

		if fGenerator.called {
			t.Fatal("expected id generator NOT to be called, but it was")
		}
		if fNotifier.called {
			t.Fatal("expected notifier NOT to be called, but it was")
		}
		if fUsers.called {
			t.Fatal("expected users NOT to be called, but it was")
		}
	}

	t.Run("chat repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fChats.err = errors.New("repo error")

		tests(t, cmdServ, fChats, fUsers, fNotifier, fGenerator)
		if fMessages.called {
			t.Fatal("expected messages NOT to be called, but it was")
		}
	})

	t.Run("message repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fMessages.err = errors.New("repo error")
		fChats.chat = chat.RehydrateChat(
			chat.ChatID(cmd.ChatID),
			chat.Name("nice name"),
			[]chat.ParticipantID{
				chat.ParticipantID(cmd.CallerID),
				chat.ParticipantID(uuid.New()),
			},
			nil,
			nil,
		)

		tests(t, cmdServ, fChats, fUsers, fNotifier, fGenerator)
		if !fMessages.called {
			t.Fatal("expected messages to be called, but it was not")
		}
	})
}

func TestHandleMarkAsRead_Success(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.ReadMessageCmd{
		ChatID:    uuid.New(),
		CallerID:  uuid.New(),
		MessageID: 1001,
	}
	fChats.chat = chat.RehydrateChat(
		chat.ChatID(cmd.ChatID),
		chat.Name("nice name"),
		[]chat.ParticipantID{
			chat.ParticipantID(cmd.CallerID),
			chat.ParticipantID(uuid.New()),
		},
		nil,
		map[chat.ParticipantID]*chat.SnowflakeID{},
	)
	fMessages.existsInChat = true

	err := cmdServ.HandleMarkAsRead(t.Context(), cmd)
	if err != nil {
		t.Fatalf("expected nil error, but got %v", err)
	}
	if !fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if !fMessages.called {
		t.Fatal("expected messages to be called, but it was not")
	}

	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}
}

func TestHandleCreateUser_InvalidInput(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	var cmd *command.CreateUserCmd
	newUserID, err := cmdServ.HandleCreateUser(t.Context(), cmd)
	if !errors.Is(err, command.ErrInvalidCommand) {
		t.Fatalf("expected wrapped error %v, but got %v", command.ErrInvalidCommand, err)
	}
	if newUserID != uuid.Nil {
		t.Fatalf("expected uuid %s (nil), but got %s", uuid.Nil, newUserID)
	}
	if fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}

	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
}
func TestHandleCreateUser_AlreadyExists(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.CreateUserCmd{
		Username: "nice username",
	}
	fUsers.err = command.ErrAlreadyExists

	newUserID, err := cmdServ.HandleCreateUser(t.Context(), cmd)
	if !errors.Is(err, command.UserAlreadyExists) {
		t.Fatalf("expected wrapped error %v, but got %s", command.UserAlreadyExists, err)
	}
	if newUserID != uuid.Nil {
		t.Fatalf("expected uuid %s (nil), but got %s", uuid.Nil, newUserID)
	}
	if !fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}

	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
}

func TestHandleCreateUser_RepoError(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	cmd := &command.CreateUserCmd{
		Username: "nice username",
	}
	fUsers.err = errors.New("repo error")

	newUserID, err := cmdServ.HandleCreateUser(t.Context(), cmd)
	if !strings.Contains(err.Error(), "handleCreateUser") {
		t.Fatalf("expected wrapped error containing 'handleCreateUser', but got %v", err)
	}
	if newUserID != uuid.Nil {
		t.Fatalf("expected uuid %s (nil), but got %s", uuid.Nil, newUserID)
	}
	if !fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}

	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
}

func TestHandleCreateUser_Success(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	userID := uuid.New()
	cmd := &command.CreateUserCmd{
		UserID:   &userID,
		Username: "nice username",
	}

	fUsers.user = user.NewUser(
		user.UserID(userID),
		user.Username(cmd.Username),
	)

	newUserID, err := cmdServ.HandleCreateUser(t.Context(), cmd)
	if err != nil {
		t.Fatalf("expected nil error, but got %v", err)
	}
	if newUserID != userID {
		t.Fatalf("expected uuid %s, but got %s", userID, newUserID)
	}
	if !fUsers.called {
		t.Fatal("expected users NOT to be called, but it was")
	}

	if fChats.called {
		t.Fatal("expected chats NOT to be called, but it was")
	}
	if fMessages.called {
		t.Fatal("expected messages NOT to be called, but it was")
	}
	if fGenerator.called {
		t.Fatal("expected id generator NOT to be called, but it was")
	}
	if fNotifier.called {
		t.Fatal("expected notifier NOT to be called, but it was")
	}
}
