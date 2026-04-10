package application

import (
	"context"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsCommandService struct {
	chats    ChatsRepository
	messages MessagesRepository
	users    UsersRepository
	notifier Notifier
}

func (cuc *ChatsCommandService) HandleCreateChat(ctx context.Context, cmd *CreateChatCmd) (*dto.Chat, error) {

	chatName, err := chat.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}

	c, err := chat.NewChat(
		idConverter[uuid.UUID, chat.ChatID](uuid.New()),
		chatName,
		idConverter[uuid.UUID, chat.ParticipantID](cmd.CallerID),
		idSliceConverter[uuid.UUID, chat.ParticipantID](cmd.Participants),
	)
	if err != nil {
		return nil, err
	}

	existingIDs, err := cuc.users.GetExistingIDs(
		ctx,
		idSliceConverter[chat.ParticipantID, user.UserID](c.Participants()),
	)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	if len(existingIDs) != len(c.Participants()) {
		diff := difference(
			existingIDs,
			idSliceConverter[chat.ParticipantID, user.UserID](c.Participants()),
		)
		return nil, &ErrUsersNotExist{
			Msg:   "some participants not exists",
			Users: diff,
		}
	}

	c, err = cuc.chats.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	go cuc.notifier.NotifyChat(c, &wsevent.NewChat{
		Chat: wsevent.Chat{
			ID:           c.ID(),
			Name:         c.Name(),
			Participants: c.Participants(),
		},
	})

	return &dto.Chat{
		ID:           idConverter[chat.ChatID, uuid.UUID](c.ID()),
		Name:         string(c.Name()),
		LastMessage:  nil,
		UnreadCount:  0,
		MessageCount: 0,
	}, nil
}
