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
		t.Fatal("expected wrapped input error")
	}
	if newChatID != uuid.Nil {
		t.Fatal("expected nil id")
	}
	if fChats.called {
		t.Fatal("chats called")
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fUsers.called {
		t.Fatal("users called")
	}
	if fNotifier.called {
		t.Fatal("users notifier called")
	}
	if fGenerator.called {
		t.Fatal("idGenerator called")
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
		t.Fatal("expected wrapped domain error")
	}
	if newChatID != uuid.Nil {
		t.Fatal("expected nil id")
	}
	if fChats.called {
		t.Fatal("chats called")
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fUsers.called {
		t.Fatal("users called")
	}
	if fNotifier.called {
		t.Fatal("notifier called")
	}
	if fGenerator.called {
		t.Fatal("idGenerator called")
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
				t.Fatal("expected error")
			}
			if newChatID != uuid.Nil {
				t.Fatal("expected nil id")
			}
			if _, ok := err.(*command.ErrUsersNotExist); !ok {
				t.Fatal("expected ErrUserNotExists error")
			}
			if typedError, ok := err.(*command.ErrUsersNotExist); ok && !slices.Equal(typedError.Users, test.expMissing) {
				t.Fatal("missing mismatch")
			}
			if !fUsers.called {
				t.Fatal("users not called")
			}
			if fChats.called {
				t.Fatal("chats called")
			}
			if fNotifier.called {
				t.Fatal("notifier called")
			}
			if fMessages.called {
				t.Fatal("messages called")
			}
			if fGenerator.called {
				t.Fatal("idGenerator called")
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
			t.Fatal("expected wrapped error, with 'create chat'")
		}
		if newChatID != uuid.Nil {
			t.Fatal("expected nil uuid")
		}
		if !fUsers.called {
			t.Fatal("users not called")
		}
		if fNotifier.called {
			t.Fatal("notifier called")
		}
		if fMessages.called {
			t.Fatal("messages called")
		}
		if fGenerator.called {
			t.Fatal("idGenerator called")
		}
	}

	t.Run("user repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fUsers.err = errors.New("repo error")

		tests(t, cmdServ, fMessages, fUsers, fNotifier, fGenerator)
		if fChats.called {
			t.Fatal("chats called")
		}
	})

	t.Run("chat repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fChats.err = errors.New("repo error")

		tests(t, cmdServ, fMessages, fUsers, fNotifier, fGenerator)
		if !fChats.called {
			t.Fatal("chats not called")
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
		t.Fatal("expected AlreadyExistsError")
	}
	if newChatID != uuid.Nil {
		t.Fatal("expected nil id")
	}
	if !fUsers.called {
		t.Fatal("users not called")
	}
	if !fChats.called {
		t.Fatal("chats not called")
	}
	if fNotifier.called {
		t.Fatal("notifier called")
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fGenerator.called {
		t.Fatal("idGenerator called")
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
		t.Fatal("expected no error")
	}
	if newChatID != uuid.UUID(fChats.chat.ID()) {
		t.Fatal("id mismatch")
	}
	if !fUsers.called {
		t.Fatal("users not called")
	}
	if !fChats.called {
		t.Fatal("chats not called")
	}
	if !fNotifier.called {
		t.Fatal("notifier not called")
	}
	if _, ok := fNotifier.event.(*wsevent.NewChat); !ok {
		t.Fatal("expected NewChat event")
	}
	if typedEvent, ok := fNotifier.event.(*wsevent.NewChat); ok {
		if typedEvent.Chat.ID != newChatID {
			t.Fatal("chat id mismatch in event")
		}
		if typedEvent.Chat.Name != cmd.Name {
			t.Fatal("name mismatch in event")
		}
		if !slices.Equal(typedEvent.Chat.Participants, cmd.Participants) {
			t.Fatal("participants mismatch in event")
		}
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fGenerator.called {
		t.Fatal("idGenerator called")
	}
}

func TestHandleCreateMessage_InvalidInput(t *testing.T) {
	cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

	var cmd *command.CreateMessageCmd
	newMessageID, err := cmdServ.HandleCreateMessage(t.Context(), cmd)
	if !errors.Is(err, command.ErrInvalidCommand) {
		t.Fatal("expected wrapped input error")
	}
	if newMessageID != 0 {
		t.Fatal("expected zero id")
	}
	if fChats.called {
		t.Fatal("chats called")
	}
	if fGenerator.called {
		t.Fatal("idGenerator called")
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fNotifier.called {
		t.Fatal("notifier called")
	}
	if fUsers.called {
		t.Fatal("users called")
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
		t.Fatal("expected NotFound error")
	}
	if newMessageID != 0 {
		t.Fatal("expected zero id")
	}
	if !fChats.called {
		t.Fatal("chats not called")
	}
	if fGenerator.called {
		t.Fatal("idGenerator called")
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fNotifier.called {
		t.Fatal("notifier called")
	}
	if fUsers.called {
		t.Fatal("users called")
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
		t.Fatal("expected domain error")
	}
	if newMessageID != 0 {
		t.Fatal("expected zero id")
	}
	if !fChats.called {
		t.Fatal("chats not called")
	}
	if !fGenerator.called {
		t.Fatal("idGenerator not called")
	}
	if fMessages.called {
		t.Fatal("messages called")
	}
	if fNotifier.called {
		t.Fatal("notifier called")
	}
	if fUsers.called {
		t.Fatal("users called")
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
			t.Fatal("expected wrapped error, with 'handleCreateMessage'")
		}
		if newMessageID != 0 {
			t.Fatal("expected zero id")
		}
		if !fChats.called {
			t.Fatal("chats not called")
		}
		if fNotifier.called {
			t.Fatal("notifier called")
		}
		if fUsers.called {
			t.Fatal("users called")
		}
	}

	t.Run("chat repo error", func(t *testing.T) {
		cmdServ, fChats, fMessages, fUsers, fNotifier, fGenerator := NewTestService()

		fChats.err = errors.New("repo error")

		tests(t, cmdServ, fChats, fUsers, fNotifier)
		if fGenerator.called {
			t.Fatal("idGenerator called")
		}
		if fMessages.called {
			t.Fatal("messages called")
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
			t.Fatal("idGenerator not called")
		}
		if !fMessages.called {
			t.Fatal("messages not called")
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
		t.Fatal("expected nil error")
	}
	if newMessageID == 0 {
		t.Fatal("expected non zero id")
	}
	if !fChats.called {
		t.Fatal("chats not called")
	}
	if !fGenerator.called {
		t.Fatal("idGenerator not called")
	}
	if !fMessages.called {
		t.Fatal("messages not called")
	}
	if !fNotifier.called {
		t.Fatal("notifier not called")
	}
	if _, ok := fNotifier.event.(*wsevent.NewMessage); !ok {
		t.Fatal("expected NewChat event")
	}
	if typedEvent, ok := fNotifier.event.(*wsevent.NewMessage); ok {
		if typedEvent.Message.ID != newMessageID {
			t.Fatal("message id mismatch in event")
		}
		if typedEvent.Message.ChatID != cmd.ChatID {
			t.Fatal("chat id mismatch in event")
		}
		if typedEvent.Message.Content != cmd.Content {
			t.Fatal("content mismatch in event")
		}
		if typedEvent.Message.SenderID != cmd.SenderID {
			t.Fatal("sender id mismatch in event")
		}
	}
	if fUsers.called {
		t.Fatal("users called")
	}
}
