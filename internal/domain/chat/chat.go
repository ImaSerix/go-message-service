package chat

import (
	"github.com/ImaSerix/go-message-service/internal/domain/message"
)

type Chat struct {
	id   ChatID
	name Name

	participants        []ParticipantID
	lastMessage         *message.Message
	lastReadedMessageID map[ParticipantID]SnowflakeID
}

func NewChat(id ChatID, name Name, callerID ParticipantID, participants []ParticipantID) (*Chat, error) {
	if len(participants) == 0 {
		return nil, ErrParticipantsEmpty
	}

	chat := Chat{
		id:                  id,
		name:                name,
		participants:        participants,
		lastReadedMessageID: make(map[ParticipantID]SnowflakeID),
	}

	if !chat.isParticipant(callerID) {
		return nil, ErrCallerIsNotParticipant
	}

	return &chat, nil
}

func RehydrateChat(id ChatID, n Name, participants []ParticipantID, lastMessage *message.Message, lastReadedMessageID map[ParticipantID]SnowflakeID) *Chat {
	return &Chat{
		id:                  id,
		name:                n,
		participants:        participants,
		lastMessage:         lastMessage,
		lastReadedMessageID: lastReadedMessageID,
	}
}

func (c *Chat) SendMessage(sender ParticipantID, message *message.Message) error {
	if !c.isParticipant(sender) {
		return ErrCallerIsNotParticipant
	}
	c.lastMessage = message

	return nil
}

// NOTE: ReadMessage ожидает, что будет передана функция, которая определит существует ли сообщение до которого хочет прочитать в чате
func (c *Chat) ReadMessage(readerID ParticipantID, messageID SnowflakeID, ExistsInChat func(SnowflakeID, ChatID) (bool, error)) error {
	if !c.isParticipant(readerID) {
		return ErrReaderIsNotParticipant
	}

	exists, err := ExistsInChat(messageID, c.id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrMessageNotInChat
	}

	if !c.lastReadedMessageID[readerID].Older(messageID) {
		return ErrNewerMessageIsAlreadyRead
	}

	c.lastReadedMessageID[readerID] = messageID
	return nil
}

func (c Chat) isParticipant(participant ParticipantID) bool {
	set := make(map[ParticipantID]struct{}, len(c.participants))
	for _, p := range c.participants {
		set[p] = struct{}{}
	}
	if _, ok := set[participant]; !ok {
		return false
	}

	return true
}

func (c Chat) Participants() []ParticipantID {
	return c.participants
}

func (c Chat) ID() ChatID {
	return c.id
}

func (c Chat) Name() Name {
	return c.name
}

func (c Chat) LastReadMessageID(readerID ParticipantID) (SnowflakeID, bool) {
	id, ok := c.lastReadedMessageID[readerID]
	return id, ok
}
