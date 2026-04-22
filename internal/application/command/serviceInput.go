package command

import (
	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/google/uuid"
)

type CreateChatInput struct {
	callerID     chat.ParticipantID
	name         chat.Name
	participants []chat.ParticipantID
}

func NewCreateChatInput(cmd *CreateChatCmd) (*CreateChatInput, error) {

	if cmd == nil {
		return nil, ErrInvalidCommand
	}

	name, err := chat.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}

	return &CreateChatInput{
		callerID:     idConverter[uuid.UUID, chat.ParticipantID](cmd.CallerID),
		name:         name,
		participants: idSliceConverter[uuid.UUID, chat.ParticipantID](cmd.Participants),
	}, nil
}

type CreateMessageInput struct {
	chatID   chat.ChatID
	senderID message.SenderID
	content  message.Content
}

func NewCreateMessageInput(cmd *CreateMessageCmd) (*CreateMessageInput, error) {

	if cmd == nil {
		return nil, ErrInvalidCommand
	}

	content, err := message.NewContent(cmd.Content)
	if err != nil {
		return nil, err
	}

	return &CreateMessageInput{
		chatID:   idConverter[uuid.UUID, chat.ChatID](cmd.ChatID),
		senderID: idConverter[uuid.UUID, message.SenderID](cmd.SenderID),
		content:  content,
	}, nil
}

type ReadMessageInput struct {
	chatID    chat.ChatID
	callerID  chat.ParticipantID
	messageID message.SnowflakeID
}

func NewReadMessageInput(cmd *ReadMessageCmd) (*ReadMessageInput, error) {
	if cmd == nil {
		return nil, ErrInvalidCommand
	}

	return &ReadMessageInput{
		chatID:    idConverter[uuid.UUID, chat.ChatID](cmd.ChatID),
		callerID:  idConverter[uuid.UUID, chat.ParticipantID](cmd.CallerID),
		messageID: message.SnowflakeID(cmd.MessageID),
	}, nil
}

type CreateUserInput struct {
	id       user.UserID
	username user.Username
}

func NewCreateUserInput(cmd *CreateUserCmd) (*CreateUserInput, error) {
	if cmd == nil {
		return nil, ErrInvalidCommand
	}

	username, err := user.NewUsername(cmd.Username)
	if err != nil {
		return nil, err
	}

	id := idConverter[uuid.UUID, user.UserID](uuid.New())
	if cmd.UserID != nil {
		id = idConverter[uuid.UUID, user.UserID](*cmd.UserID)
	}

	return &CreateUserInput{
		id:       id,
		username: username,
	}, nil
}
