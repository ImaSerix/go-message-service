package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsComandService struct {
	chats    ChatsRepository
	messages MessagesRepository
	users    UsersRepository
	notifier Notifier
}

func (cuc *ChatsComandService) HandleCreateChat(ctx context.Context, cmd *CreateChatCmd) (*dto.Chat, error) {

	chatName, err := chat.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}
	chatParticipants := make([]chat.ParticipantID, len(cmd.Participants))
	for _, p := range cmd.Participants {
		chatParticipants = append(chatParticipants, chat.ParticipantID(p))
	}

	chat, err := chat.NewChat(chat.ChatID(uuid.New()), chatName, chat.ParticipantID(cmd.CallerID), chatParticipants)
	if err != nil {
		return nil, err
	}

	userIDs := make([]user.UserID, len(chat.Participants()))
	for _, p := range chat.Participants() {
		userIDs = append(userIDs, user.UserID(p))
	}
	existingIDs, err := cuc.users.GetExistingIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	if len(existingIDs) != len(userIDs) {
		diff := difference(userIDs, existingIDs)
		return nil, &ErrUsersNotExists{
			Msg:   "some participants not exists",
			Users: diff,
		}
	}

	chat, err = cuc.chats.Create(ctx, chat)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create chat: %w", err)
	}

	go cuc.notifier.NotifyChat(ctx, chat.ID(), &wsevent.NewChat{
		Chat: wsevent.Chat{
			ID:           chat.ID(),
			Name:         chat.Name(),
			Participants: chat.Participants(),
		},
	})

	return &dto.Chat{
		ID:           uuid.UUID(chat.ID()),
		Name:         string(chat.Name()),
		LastMessage:  nil,
		UnreadCount:  0,
		MessageCount: 0,
	}, nil
}
