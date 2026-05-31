package command_test

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type fakeChatsRepo struct {
	chat   *chat.Chat
	err    error
	called bool
}

func (r *fakeChatsRepo) Create(ctx context.Context, chat *chat.Chat) (*chat.Chat, error) {
	r.called = true
	return r.chat, r.err
}
func (r *fakeChatsRepo) LoadForSend(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	r.called = true
	return r.chat, r.err
}
func (r *fakeChatsRepo) LoadForRead(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	r.called = true
	return r.chat, r.err
}

type fakeMessagesRepo struct {
	message      *message.Message
	existsInChat bool
	err          error
	called       bool
}

func (r *fakeMessagesRepo) Create(ctx context.Context, message *message.Message) (*message.Message, error) {
	r.called = true
	return r.message, r.err
}
func (r *fakeMessagesRepo) ExistsInChat(ctx context.Context, messageID message.SnowflakeID, chatID chat.ChatID) (bool, error) {
	r.called = true
	return r.existsInChat, r.err
}
func (r *fakeMessagesRepo) UpdateLastReaded(ctx context.Context, readerID chat.ParticipantID, chat *chat.Chat) error {
	r.called = true
	return r.err
}

type fakeUsersRepo struct {
	user        *user.User
	existingIDs []user.UserID
	err         error
	called      bool
}

func (r *fakeUsersRepo) GetExistingIDs(ctx context.Context, ids []user.UserID) ([]user.UserID, error) {
	r.called = true
	return r.existingIDs, r.err
}
func (r *fakeUsersRepo) Create(ctx context.Context, user *user.User) (*user.User, error) {
	r.called = true
	return r.user, r.err
}

type fakeNotifier struct {
	called bool
	event  wsevent.WSEvent
}

func (n *fakeNotifier) SendToUsers(participants []uuid.UUID, event wsevent.WSEvent) {
	n.called = true
	n.event = event
}

type fakeIDGenerator struct {
	called bool
	id     message.SnowflakeID
}

func (g *fakeIDGenerator) NewID() message.SnowflakeID {
	g.called = true
	return g.id
}

type syncRunner struct{}

func (r *syncRunner) Run(fn func()) {
	fn()
}
