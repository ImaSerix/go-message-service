package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type CommandService struct {
	chats     ChatsRepository
	messages  MessagesRepository
	users     UsersRepository
	notifier  Notifier
	generator IDGenerator
	runner    Runner
}

func NewCommandService(chats ChatsRepository, messages MessagesRepository, users UsersRepository, notifier Notifier, generator IDGenerator, runner Runner) *CommandService {
	return &CommandService{
		chats:     chats,
		messages:  messages,
		users:     users,
		notifier:  notifier,
		generator: generator,
		runner:    runner,
	}
}

func (cs *CommandService) HandleCreateChat(ctx context.Context, cmd *CreateChatCmd) (uuid.UUID, error) {

	in, err := NewCreateChatInput(cmd)
	if err != nil {
		return uuid.Nil, err
	}

	c, err := chat.NewChat(
		idConverter[uuid.UUID, chat.ChatID](uuid.New()),
		in.name,
		in.callerID,
		in.participants,
	)
	if err != nil {
		return uuid.Nil, err
	}

	existingIDs, err := cs.users.GetExistingIDs(
		ctx,
		idSliceConverter[chat.ParticipantID, user.UserID](c.Participants()),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create chat: %w", err)
	}

	miss := missing(
		idSliceConverter[chat.ParticipantID, user.UserID](c.Participants()),
		existingIDs,
	)

	if len(miss) > 0 {

		return uuid.Nil, &ErrUsersNotExist{
			Msg:   "some participants not exists",
			Users: miss,
		}
	}

	c, err = cs.chats.Create(ctx, c)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return uuid.Nil, ErrAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("create chat: %w", err)
	}

	cs.runner.Run(func() {
		cs.notifier.SendToUsers(idSliceConverter[chat.ParticipantID, uuid.UUID](c.Participants()), &wsevent.NewChat{
			Chat: wsevent.Chat{
				ID:           idConverter[chat.ChatID, uuid.UUID](c.ID()),
				Name:         string(c.Name()),
				Participants: idSliceConverter[chat.ParticipantID, uuid.UUID](c.Participants()),
			},
		})
	})

	return idConverter[chat.ChatID, uuid.UUID](c.ID()), nil
}

func (cs *CommandService) HandleCreateMessage(ctx context.Context, cmd *CreateMessageCmd) (int64, error) {

	in, err := NewCreateMessageInput(cmd)
	if err != nil {
		return 0, err
	}

	c, err := cs.chats.LoadForSend(ctx, in.chatID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return 0, ErrChatNotExists
		default:
			return 0, fmt.Errorf("handleCreateMessage: %w", err)
		}
	}

	id := cs.generator.NewID()
	m := message.NewMessage(
		id,
		idConverter[chat.ChatID, message.ChatID](c.ID()),
		in.senderID,
		in.content,
	)

	if err := c.SendMessage(idConverter[message.SenderID, chat.ParticipantID](in.senderID), m); err != nil {
		return 0, err
	}

	m, err = cs.messages.Create(ctx, m)
	if err != nil {
		return 0, fmt.Errorf("handleCreateMessage: %w", err)
	}

	cs.runner.Run(func() {
		cs.notifier.SendToUsers(idSliceConverter[chat.ParticipantID, uuid.UUID](c.Participants()), &wsevent.NewMessage{
			Message: wsevent.Message{
				ID:        int64(m.ID()),
				ChatID:    idConverter[message.ChatID, uuid.UUID](m.ChatID()),
				SenderID:  idConverter[message.SenderID, uuid.UUID](m.SenderID()),
				Content:   string(m.Content()),
				CreatedAt: *m.CreatedAt(),
			},
		})
	})

	return int64(m.ID()), nil
}

func (cs *CommandService) HandleMarkAsRead(ctx context.Context, cmd *ReadMessageCmd) error {

	in, err := NewReadMessageInput(cmd)
	if err != nil {
		return err
	}

	c, err := cs.chats.LoadForRead(ctx, in.chatID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return ErrChatNotExists
		default:
			return fmt.Errorf("markAsRead: %w", err)
		}
	}

	err = c.ReadMessage(
		in.callerID,
		chat.SnowflakeID(in.messageID),
		func(si chat.SnowflakeID, ci chat.ChatID) (bool, error) {
			exists, err := cs.messages.ExistsInChat(ctx, in.messageID, in.chatID)
			if err != nil {
				return false, fmt.Errorf("markAsRead: %w", err)
			}
			return exists, nil
		},
	)

	if err != nil {
		return err
	}

	if err := cs.messages.UpdateLastReaded(ctx, in.callerID, c); err != nil {
		return fmt.Errorf("markAsRead: %w", err)
	}
	return nil
}

func (cs *CommandService) HandleCreateUser(ctx context.Context, cmd *CreateUserCmd) (uuid.UUID, error) {

	in, err := NewCreateUserInput(cmd)
	if err != nil {
		return uuid.Nil, err
	}

	u := user.NewUser(in.id, in.username)

	u, err = cs.users.Create(ctx, u)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return uuid.Nil, UserAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("handleCreateUser: %w", err)
	}

	return idConverter[user.UserID, uuid.UUID](u.ID()), nil
}
